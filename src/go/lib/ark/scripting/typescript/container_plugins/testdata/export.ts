/*
This file is used to test that a plugin was executed and returned a k8s manifest
 */
import microservice from "ark/plugins/@mantl/sre/microservice"

const manifest = microservice({})

export default manifest