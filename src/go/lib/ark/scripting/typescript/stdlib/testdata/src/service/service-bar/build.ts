import * as arksdk from 'arksdk'
import * as filepath from 'arksdk/filepath'

export const image = arksdk.actions.buildDockerImage({
    name: "test",
    attributes: {
        repo: "gcr.io/managed-infrastructure",
        dockerfile: filepath.load("./Dockerfile"),
    },
    sourceFiles: filepath.glob("**/*.ts"),
})