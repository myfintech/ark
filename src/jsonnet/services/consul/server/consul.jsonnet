local configMap = import './config-map.jsonnet';
local disruptionBudget = import './disruption-budget.jsonnet';
local role = import './role.jsonnet';
local serviceAccount = import './service-account.jsonnet';
local service = import './service.jsonnet';
local statefulSet = import './stateful-set.jsonnet';


function() {
  local consulDisruptionBudget = disruptionBudget(),
  local consulServiceAccount = serviceAccount(),
  local consulRole = role(),
  local consulConfigMap = configMap(),
  local consulService = service(),
  local consulStatefulSet = statefulSet(
    serviceAccountName=consulServiceAccount.serverServiceAccount.metadata.name,
    serverConfigMap=consulConfigMap.serverConfigMap.metadata.name,
  ),

  manifest():: [
    consulDisruptionBudget.disruptionBudget,
    consulServiceAccount.serverServiceAccount,
    consulRole.serverRole,
    consulRole.serverRoleBinding,
    consulConfigMap.serverConfigMap,
    consulService.serverService,
    consulService.uiService,
    consulStatefulSet,
  ],
}
