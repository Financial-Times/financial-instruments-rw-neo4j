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

const writeStatement = `MERGE (n:Thing{uuid: {uuid}})
			set n={props}
			set n :Concept
			set n :FinancialInstrument
			set n :Equity`

const deleteEntityRelationshipsStatement = `MATCH (t:Thing {uuid:{uuid}})
						OPTIONAL MATCH (i:Identifier)-[ir:IDENTIFIES]->(t)
						DELETE ir, i`

const countStatement = `MATCH (n:FinancialInstrument) return count(n) as count`

const organizationRelationshipStatement = `MERGE (fi:Thing {uuid: {uuid}})
					MERGE (orgUpp:Identifier:UPPIdentifier{value:{orgUuid}})
					MERGE (orgUpp)-[:IDENTIFIES]->(o:Thing) ON CREATE SET o.uuid = {orgUuid}
					MERGE (fi)-[:ISSUED_BY]->(o)`

const clearIdentifierStatement = `MATCH (p:Thing {uuid: {uuid}})
				OPTIONAL MATCH (p)<-[ir:IDENTIFIES]-(i:Identifier)
				REMOVE p:Concept
				REMOVE p:FinancialInstrument
				REMOVE p:Equity
				DELETE ir, i
				SET p={props}`

const removeUnusedNodeStatement = `MATCH (p:Thing {uuid: {uuid}})
				OPTIONAL MATCH (p)-[a]-(x)
				WITH p, count(a) AS relCount
				WHERE relCount = 0
				DELETE p`

const idsStatement = `MATCH (fi:FinancialInstrument) RETURN fi.uuid as id, fi.hash as hash SKIP {skip} LIMIT {limit}`