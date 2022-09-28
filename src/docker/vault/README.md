Developer Vault
===

> *NOTE:* This should not be used in a production setting. This is for development only.

# Overview
This is a pre-initialized vault which automatically unseals itself on boot for use in development. 

# Why?
There have been numerous [Github issues][1] opened for Hashicorp Vault to implement an easier method of developing locally.
While there is interest in the community, vault is a security product and therefore probably shouldn't implement functionality that is inherently insecure.

To get around this we have created a pre-initialized vault container with a deterministic unseal key and root token.

# Unseal config 
This configuration has been committed to this image under `/opt/vault/unseal.json`
There is no need to unseal this manually, the container performs this operation automatically on start.

| name | description | value |
| ---- | ----------- | ----- |
| root token | The `-dev` initialized root token | `root`
| unseal key | The base64 encoded unseal key used by `vault operator unseal` | `AQuQ24FFO5cVMD85l1Is5hT7uilDIa+Z+CjLp/P3Qno=` |

# Building the image
```shell
ark run dev_vault.build.image
```

# Try it out using docker
This will demonstrate that this vault container automatically unseals itself.
Using the provided `root_token` above you can visit http://localhost:8200 and log in.
After logging in and creating a secret you should be able to stop the container, boot again, and your new secret should still be present. Checking the logs in the container you'll see a line `/mnt/vault already exists, skipping initialization`.
```shell
mkdir -p /tmp/vault && \
  docker run -v /tmp/vault:/mnt/vault -p 8200:8200 gcr.io/[insert-google-project]/ark/dev/vault:latest
```

# Persistent Data
This image uses a vaults file storage engine.
When using docker or kubernetes you will need to mount a volume at `/mnt/vault`.
The `entrypoint.sh` will follow this procedure to prepare the container and its data.
- Copy `VAULT_BASE_STORAGE` to`VAULT_DATA_VOLUME` if a `data` directory doesn't exist at this path.
- Start `vault server -config=/opt/vault/config.hcl`
- Probe for `VAULT_ADDR`
- Unseal using `/opt/vault/unseal.json`

# Environmental Variables
You shouldn't need to modify any of these values unless you're feeling adventurous.

| name | description | default |
| ---- | ----------- | ------- |
| `VAULT_DATA_VOLUME` | The volume mounting location (use this in k8s PVs) | /mnt/vault |
| `VAULT_BASE_STORAGE` | The location which contains the pre-initialized vault | /opt/vault/data |
| `VAULT_ADDR` | The default listener in vault. To change this default modify config.hcl | http://localhost:8200 |


[1]: https://github.com/hashicorp/vault/issues/1160
