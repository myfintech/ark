#!/bin/bash -e


sdm update

if [ -n "$SDM_SERVICE_TOKEN" ];
then
        sdm --admin-token $SDM_SERVICE_TOKEN login
        echo "Signed in with service account"
else
        cp -a /tmp/sdm /root/.sdm
        echo "Service account not available. Defaulting to user provided sdm credentials"
fi

sdm listen
