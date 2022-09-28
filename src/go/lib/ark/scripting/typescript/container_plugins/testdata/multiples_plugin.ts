/*
Test the import of multiple plugins
 */
import microservice from "ark/plugins/@ark/sre/microservice"
import vaultK8sSA from "ark/plugins/@ark/sre/vault-service-account"

microservice({})
vaultK8sSA({})
