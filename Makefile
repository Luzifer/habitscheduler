default: apidocs

apidocs:
		yaml2json < apidocs.yaml > apidocs.json
		HASH=`ipfs add apidocs.json | awk '{print $$2}'`; \
		sed "s|url=Qm.*)|url=$$HASH)|" README.md > README.tmp
		mv README.tmp README.md

