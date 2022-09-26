local cConfigMap = import './client/config-map.jsonnet';
local cDaemonSet = import './client/daemon-set.jsonnet';
local cRole = import './client/role.jsonnet';
local cServiceAccount = import './client/service-account.jsonnet';
local dService = import './dns-service.jsonnet';
local sConfigMap = import './server/config-map.jsonnet';
local sDisruptionBudget = import './server/disruption-budget.jsonnet';
local sRole = import './server/role.jsonnet';
local sServiceAccount = import './server/service-account.jsonnet';
local sService = import './server/service.jsonnet';
local sStatefulSet = import './server/stateful-set.jsonnet';
local wDeployment = import './webhook/deployment.jsonnet';
local wMutatingWebhook = import './webhook/mutating-webhook.jsonnet';
local wConfig = import './webhook/mutating-webhook.jsonnet';
local wRole = import './webhook/role.jsonnet';
local wServiceAccount = import './webhook/service-account.jsonnet';
local wService = import './webhook/service.jsonnet';


function() {
  local clientServiceAccount = cServiceAccount(),
  local clientConfigMap = cConfigMap(),
  local clientRole = cRole(),
  local serverDisruptionBudget = sDisruptionBudget(),
  local serverServiceAccount = sServiceAccount(),
  local serverConfigMap = sConfigMap(),
  local serverRole = sRole(),
  local webhookServiceAccount = wServiceAccount(),
  local webhookRole = wRole(),
  local serverService = sService(),
  local webhookService = wService(),
  local webhookConfig = wConfig(),
  local dnsService = dService(),
  local clientDaemonSet = cDaemonSet(
    serviceAccountName=clientServiceAccount.clientServiceAccount.metadata.name,
    clientConfigMap=clientConfigMap.clientConfigMap.metadata.name,
  ),
  local serverStatefulSet = sStatefulSet(
    serviceAccountName=serverServiceAccount.serverServiceAccount.metadata.name,
    serverConfigMap=serverConfigMap.serverConfigMap.metadata.name,
  ),
  local webhookDeployment = wDeployment(
    serviceAccountName=webhookServiceAccount.webhookServiceAccount.metadata.name,
  ),

  manifest():: [
    serverDisruptionBudget.disruptionBudget,
    clientServiceAccount.clientServiceAccount,
    serverServiceAccount.serverServiceAccount,
    webhookServiceAccount.webhookServiceAccount,
    clientConfigMap.clientConfigMap,
    serverConfigMap.serverConfigMap,
    clientRole.clientRole,
    clientRole.clientRoleBinding,
    serverRole.serverRole,
    serverRole.serverRoleBinding,
    webhookRole.webhookRole,
    webhookRole.webhookRoleBinding,
    webhookService.webhookService,
    serverService.serverService,
    serverService.uiService,
    clientDaemonSet,
    serverStatefulSet,
    webhookDeployment,
    webhookConfig.mutatingWebhook,
    dnsService.consulDNSService,
  ],
}
