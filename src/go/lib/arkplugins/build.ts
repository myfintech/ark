import * as ark from "arksdk"

//FIXME(rckgomz): find a better way to do this.
import ghr from "./cicd-gha-runner/build"
import consul from "./consul/build"
import coreProxy from "./core-proxy/build"
import datadog from "./datadog/build"
import gcloudEmulator from "./gcloud-emulator/build"
import kafka from "./kafka/build"
import kubeState from "./kube-state/build"
import microservice from "./microservice/build"
import nsReaper from "./ns-reaper/build"
import postgres from "./postgres/build"
import redis from "./redis/build"
import sdm from "./sdm/build"
import terraformCloudAgent from "./terraform-cloud-agent/build"
import vault from "./vault/build"
import vaultServiceAccount from "./vault-service-account/build"

ark.actions.group({
    name: "build.all",
    dependsOn: [
        ghr,
        consul,
        coreProxy,
        datadog,
        gcloudEmulator,
        kafka,
        kubeState,
        microservice,
        nsReaper,
        postgres,
        redis,
        sdm,
        terraformCloudAgent,
        vault,
        vaultServiceAccount,
    ]
})