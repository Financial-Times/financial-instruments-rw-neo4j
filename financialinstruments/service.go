package financialinstruments

import (
	"encoding/json"
	"fmt"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/Financial-Times/up-rw-app-api-go/rwapi"
	"github.com/jmcvetta/neoism"
)

type service struct {
	conn neoutils.NeoConnection
}

const batchSize = 4096

//NewCypherFinancialInstrumentService returns a new service responsible for writing financial instruments in Neo4j
func NewCypherFinancialInstrumentService(cypherRunner neoutils.NeoConnection) service {
	return service{cypherRunner}
}

func (s service) Initialise() error {
	err := s.conn.EnsureIndexes(map[string]string{
		"Identifier": "value",
	})

	if err != nil {
		return err
	}

	return s.conn.EnsureConstraints(map[string]string{
		"Thing":               "uuid",
		"Concept":             "uuid",
		"FinancialInstrument": "uuid",
		"UPPIdentifier":       "value",
		"FactsetIdentifier":   "value",
		"FIGIIdentifier":      "value",
	})
}

func (s service) Read(uuid string, transactionID string) (interface{}, bool, error) {

	results := []financialInstrument{}

	readQuery := &neoism.CypherQuery{
		Statement: `MATCH (fi:FinancialInstrument {uuid:{uuid}})
				OPTIONAL MATCH (fi)-[:ISSUED_BY]->(org:Thing)
				OPTIONAL MATCH (upp:UPPIdentifier)-[:IDENTIFIES]->(fi)
				OPTIONAL MATCH (factset:FactsetIdentifier)-[:IDENTIFIES]->(fi)
				OPTIONAL MATCH (figi:FIGIIdentifier)-[:IDENTIFIES]->(fi)
				OPTIONAL MATCH (wsod:WSODIdentifier)-[:IDENTIFIES]->(fi)
				return fi.uuid as uuid,
					fi.prefLabel as prefLabel,
					org.uuid as issuedBy,
					{uuids:collect(distinct upp.value),
					figiCode:figi.value,
					factsetIdentifier:factset.value,
					wsodIdentifier: wsod.value} as alternativeIdentifiers`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
		Result: &results,
	}

	if err := s.conn.CypherBatch([]*neoism.CypherQuery{readQuery}); err != nil || len(results) == 0 {
		return financialInstrument{}, false, err
	}

	return results[0], true, nil

}

func createNewIdentifierQuery(uuid string, identifierLabel string, identifierValue string) *neoism.CypherQuery {
	statementTemplate := fmt.Sprintf(`MERGE (t:Thing {uuid:{uuid}})
				CREATE (i:Identifier {value:{value}})
				MERGE (t)<-[:IDENTIFIES]-(i)
				set i : %s`, identifierLabel)

	query := &neoism.CypherQuery{
		Statement: statementTemplate,
		Parameters: map[string]interface{}{
			"uuid":  uuid,
			"value": identifierValue,
		},
	}
	return query
}

func getNewIdentifierQueries(fi financialInstrument) []*neoism.CypherQuery {
	queries := []*neoism.CypherQuery{}

	//ADD all the IDENTIFIER nodes and IDENTIFIES relationships
	for _, alternativeUUID := range fi.AlternativeIdentifiers.UUIDS {
		if alternativeUUID != "" {
			queries = append(queries, createNewIdentifierQuery(fi.UUID, uppIdentifierLabel, alternativeUUID))
		}
	}

	if fi.AlternativeIdentifiers.FactsetIdentifier != "" {
		queries = append(queries, createNewIdentifierQuery(fi.UUID, factsetIdentifierLabel, fi.AlternativeIdentifiers.FactsetIdentifier))
	}

	if fi.AlternativeIdentifiers.FIGICode != "" {
		queries = append(queries, createNewIdentifierQuery(fi.UUID, figiIdentifierLabel, fi.AlternativeIdentifiers.FIGICode))
	}

	if fi.AlternativeIdentifiers.WSODIdentifier != "" {
		queries = append(queries, createNewIdentifierQuery(fi.UUID, wsodIdentifierLabel, fi.AlternativeIdentifiers.WSODIdentifier))
	}

	return queries
}

func (s service) Write(thing interface{}, transactionID string) error {

	hash, err := writeHash(thing)
	if err != nil {
		return err
	}

	fi := thing.(financialInstrument)

	params := map[string]interface{}{
		"uuid": fi.UUID,
		"hash": hash,
	}

	if fi.PrefLabel != "" {
		params["prefLabel"] = fi.PrefLabel
	}

	queries := []*neoism.CypherQuery{}

	deleteEntityRelationshipsQuery := &neoism.CypherQuery{
		Statement: `MATCH (t:Thing {uuid:{uuid}})
				OPTIONAL MATCH (t)-[is:ISSUED_BY]->(org:Thing)
				OPTIONAL MATCH (i:Identifier)-[ir:IDENTIFIES]->(t)
				DELETE ir, is, i`,
		Parameters: map[string]interface{}{
			"uuid": fi.UUID,
		},
	}
	queries = append(queries, deleteEntityRelationshipsQuery)

	writeQuery := &neoism.CypherQuery{
		Statement: `MERGE (t:Thing{uuid: {uuid}})
			set t={props}
			set t :Concept
			set t :FinancialInstrument`,
		Parameters: map[string]interface{}{
			"uuid":  fi.UUID,
			"props": params,
		},
	}
	queries = append(queries, writeQuery)
	queries = append(queries, getNewIdentifierQueries(fi)...)

	if fi.IssuedBy != "" {
		orgUUID := fi.IssuedBy

		orgResults := []struct {
			UUID string `json:"uuid"`
		}{}

		findOrganisationQuery := &neoism.CypherQuery{
			Statement: `MATCH (i:Identifier {value: {uuid}})-[:IDENTIFIES]->(org:Thing) RETURN org.uuid as uuid`,
			Parameters: map[string]interface{}{
				"uuid": fi.IssuedBy,
			},
			Result: &orgResults,
		}

		if err := s.conn.CypherBatch([]*neoism.CypherQuery{findOrganisationQuery}); err != nil {
			fmt.Println(err)
			return err
		}

		if len(orgResults) > 0 {
			orgUUID = orgResults[0].UUID
		}

		organizationRelationshipQuery := &neoism.CypherQuery{
			Statement: `MERGE (fi:Thing {uuid: {uuid}})
					MERGE (orgUpp:Identifier:UPPIdentifier{value:{orgUuid}})
					MERGE (orgUpp)-[:IDENTIFIES]->(o:Thing) ON CREATE SET o.uuid = {orgUuid}
					MERGE (fi)-[:ISSUED_BY]->(o)`,
			Parameters: map[string]interface{}{
				"uuid":    fi.UUID,
				"orgUuid": orgUUID,
			},
		}
		queries = append(queries, organizationRelationshipQuery)
	}

	return s.conn.CypherBatch(queries)
}

func (s service) Delete(uuid string, transactionID string) (bool, error) {
	clearNode := &neoism.CypherQuery{
		Statement: `MATCH (t:Thing {uuid: {uuid}})
				OPTIONAL MATCH (t)-[is:ISSUED_BY]->(org:Thing)
				OPTIONAL MATCH (t)<-[ir:IDENTIFIES]-(i:Identifier)
				REMOVE t:Concept:FinancialInstrument
				DELETE is, ir, i
				SET t={props}`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
			"props": map[string]interface{}{
				"uuid": uuid,
			},
		},
		IncludeStats: true,
	}

	removeNodeIfUnused := &neoism.CypherQuery{
		Statement: `MATCH (t:Thing {uuid: {uuid}})
				OPTIONAL MATCH (t)-[a]-(x)
				WITH t, count(a) AS relCount
				WHERE relCount = 0
				DELETE t`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
	}

	err := s.conn.CypherBatch([]*neoism.CypherQuery{clearNode, removeNodeIfUnused})

	stats, err := clearNode.Stats()
	if err != nil {
		return false, err
	}

	var deleted bool
	if stats.ContainsUpdates && stats.LabelsRemoved > 0 {
		deleted = true
	}

	return deleted, err
}

func (s service) Count() (int, error) {
	results := []struct {
		Count int `json:"count"`
	}{}

	query := &neoism.CypherQuery{
		Statement: `MATCH (fi:FinancialInstrument) return count(fi) as count`,
		Result:    &results,
	}
	err := s.conn.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return 0, err
	}

	return results[0].Count, nil
}

func (s service) DecodeJSON(dec *json.Decoder) (interface{}, string, error) {
	fi := financialInstrument{}
	err := dec.Decode(&fi)
	return fi, fi.UUID, err
}

func (s service) IDs(f func(id rwapi.IDEntry) (bool, error)) error {

	for skip := 0; ; skip += batchSize {
		results := []rwapi.IDEntry{}
		readQuery := &neoism.CypherQuery{
			Statement: `MATCH (fi:FinancialInstrument) RETURN fi.uuid as id, fi.hash as hash SKIP {skip} LIMIT {limit}`,
			Parameters: map[string]interface{}{
				"limit": batchSize,
				"skip":  skip,
			},
			Result: &results,
		}

		if err := s.conn.CypherBatch([]*neoism.CypherQuery{readQuery}); err != nil {
			return nil
		}
		if len(results) == 0 {
			return nil
		}
		for _, result := range results {
			more, err := f(result)
			if !more || err != nil {
				return err
			}
		}

	}
}

func (s service) Check() error {
	return neoutils.Check(s.conn)
}
