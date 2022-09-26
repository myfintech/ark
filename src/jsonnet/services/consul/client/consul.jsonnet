local configMap = import './config-map.jsonnet';
local daemonSet = import './daemon-set.jsonnet';
local role = import './role.jsonnet';
local serviceAccount = import './service-account.jsonnet';

function() {
  local consulServiceAccount = serviceAccount(),
  local consulRole = role(),
  local consulConfigMap = configMap(),
  local consulDaemonSet = daemonSet(
    serviceAccountName=consulServiceAccount.clientServiceAccount.metadata.name,
    clientConfigMap=consulConfigMap.clientConfigMap.metadata.name,
  ),

  manifest():: [
    consulServiceAccount.clientServiceAccount,
    consulRole.clientRole,
    consulRole.clientRoleBinding,
    consulConfigMap.clientConfigMap,
    consulDaemonSet,
  ],
}
