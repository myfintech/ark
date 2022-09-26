import * as filepath from "arksdk/filepath"
import {buildPlugin} from "ark/external/plugin_builder"

const currentFolderName = filepath.getFolderNameFromCurrentLocation()
export default buildPlugin({
    pluginName: currentFolderName,
    imageName: currentFolderName,
})


