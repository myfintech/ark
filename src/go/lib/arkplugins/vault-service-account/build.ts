import * as filepath from "arksdk/filepath"
import {buildPlugin} from "ark/external/plugin_builder"

export default buildPlugin({
    pluginName: filepath.getFolderNameFromCurrentLocation(),
    imageName: 'vault-k8s-sa',
})


