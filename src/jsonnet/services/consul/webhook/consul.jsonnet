local deployment = import './deployment.jsonnet';
local webhook = import './mutating-webhook.jsonnet';
local role = import './role.jsonnet';
local serviceAccount = import './service-account.jsonnet';
local service = import './service.jsonnet';


function() {
  local consulServiceAccount = serviceAccount(),
  local consulRole = role(),
  local consulService = service(),
  local consulWebhook = webhook(),
  local consulDeployment = deployment(
    serviceAccountName=consulServiceAccount.webhookServiceAccount.metadata.name,
  ),

  manifest():: [
    consulServiceAccount.webhookServiceAccount,
    consulRole.webhookRole,
    consulRole.webhookRoleBinding,
    consulService.webhookService,
    consulWebhook.mutatingWebhook,
    consulDeployment,
  ],
}
