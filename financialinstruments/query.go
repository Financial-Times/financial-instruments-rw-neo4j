package financialinstruments

const readByUUIDQuery = `MATCH (fi:FinancialInstrument {uuid:{uuid}})
				OPTIONAL MATCH (fi)-[:ISSUED_BY]->(org:Thing)
				OPTIONAL MATCH (upp:UPPIdentifier)-[:IDENTIFIES]->(fi)
				OPTIONAL MATCH (factset:FactsetIdentifier)-[:IDENTIFIES]->(fi)
				OPTIONAL MATCH (figi:FIGIIdentifier)-[:IDENTIFIES]->(fi)
				return fi.uuid as uuid,
					fi.prefLabel as prefLabel,
					org.uuid as issuedBy,
					{uuids:collect(distinct upp.value),
					figiCode:figi.value,
					factsetIdentifier:factset.value} as alternativeIdentifiers`

const newIdentifierQuery = `MERGE (t:Thing {uuid:{uuid}})
				CREATE (i:Identifier {value:{value}})
				MERGE (t)<-[:IDENTIFIES]-(i)
				set i : %s`

const writeQuery = `MERGE (n:Thing{uuid: {uuid}})
			set n={props}
			set n :Concept
			set n :FinancialInstrument
			set n :Equity`

const deleteEntityRelationshipsQuery = `MATCH (t:Thing {uuid:{uuid}})
						OPTIONAL MATCH (i:Identifier)-[ir:IDENTIFIES]->(t)
						DELETE ir, i`

const countQuery = `MATCH (n:FinancialInstrument) return count(n) as count`

const organizationRelationshipQuery = `MERGE (fi:Thing {uuid: {uuid}})
					MERGE (orgUpp:Identifier:UPPIdentifier{value:{orgUuid}})
					MERGE (orgUpp)-[:IDENTIFIES]->(o:Thing) ON CREATE SET o.uuid = {orgUuid}
					MERGE (fi)-[:ISSUED_BY]->(o)`

const clearIdentifierQuery = `MATCH (p:Thing {uuid: {uuid}})
				OPTIONAL MATCH (p)<-[ir:IDENTIFIES]-(i:Identifier)
				REMOVE p:Concept
				REMOVE p:FinancialInstrument
				REMOVE p:Equity
				DELETE ir, i
				SET p={props}`

const removeUnusedNodeQuery = `MATCH (p:Thing {uuid: {uuid}})
				OPTIONAL MATCH (p)-[a]-(x)
				WITH p, count(a) AS relCount
				WHERE relCount = 0
				DELETE p`