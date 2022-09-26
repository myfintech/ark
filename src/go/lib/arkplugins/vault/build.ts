import * as filepath from 'arksdk/filepath'
import {buildPlugin} from "ark/external/plugin_builder"
import {image} from "../../../../docker/vault/build";

const currentFolderName = filepath.getFolderNameFromCurrentLocation()
export default buildPlugin({
    pluginName: currentFolderName,
    imageName: currentFolderName,
    env: {
        IMAGE_URL: image.attributes.url
    },
    dependsOn: [image]
})
