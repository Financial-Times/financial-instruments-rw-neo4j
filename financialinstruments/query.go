package financialinstruments

const readByUUIDStatement = `MATCH (fi:FinancialInstrument {uuid:{uuid}})
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
					wsodIdentifier: wsod.value} as alternativeIdentifiers`

const newIdentifierStatement = `MERGE (t:Thing {uuid:{uuid}})
				CREATE (i:Identifier {value:{value}})
				MERGE (t)<-[:IDENTIFIES]-(i)
				set i : %s`

const writeStatement = `MERGE (t:Thing{uuid: {uuid}})
			set t={props}
			set t :Concept
			set t :FinancialInstrument
			set t :Equity`

const deleteEntityRelationshipsStatement = `MATCH (t:Thing {uuid:{uuid}})
						OPTIONAL MATCH (t)-[is:ISSUED_BY]->(org:Thing)
						OPTIONAL MATCH (i:Identifier)-[ir:IDENTIFIES]->(t)
						DELETE ir, is, i`

const countStatement = `MATCH (fi:FinancialInstrument) return count(fi) as count`

const organizationRelationshipStatement = `MERGE (fi:Thing {uuid: {uuid}})
					MERGE (orgUpp:Identifier:UPPIdentifier{value:{orgUuid}})
					MERGE (orgUpp)-[:IDENTIFIES]->(o:Thing) ON CREATE SET o.uuid = {orgUuid}
					MERGE (fi)-[:ISSUED_BY]->(o)`

const clearNodeStatement = `MATCH (t:Thing {uuid: {uuid}})
				OPTIONAL MATCH (t)-[is:ISSUED_BY]->(org:Thing)
				OPTIONAL MATCH (t)<-[ir:IDENTIFIES]-(i:Identifier)
				REMOVE t:Concept:FinancialInstrument:Equity
				DELETE is, ir, i
				SET t={props}`

const removeUnusedNodeStatement = `MATCH (t:Thing {uuid: {uuid}})
				OPTIONAL MATCH (t)-[a]-(x)
				WITH t, count(a) AS relCount
				WHERE relCount = 0
				DELETE t`

const idsStatement = `MATCH (fi:FinancialInstrument) RETURN fi.uuid as id, fi.hash as hash SKIP {skip} LIMIT {limit}`

const findOrganisationStatement = `MATCH (i:Identifier {value: {uuid}})-[:IDENTIFIES]->(org:Thing) RETURN org.uuid as uuid`