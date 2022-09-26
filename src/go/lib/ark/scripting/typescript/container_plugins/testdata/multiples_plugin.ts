/*
Test the import of multiple plugins
 */
import microservice from "ark/plugins/@mantl/sre/microservice"
import vaultK8sSA from "ark/plugins/@mantl/sre/vault-service-account"

microservice({})
vaultK8sSA({})