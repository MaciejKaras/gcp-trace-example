#!/bin/bash

set -x

if [[ -z "$TF_VAR_project_id" ]]; then
  echo -e "Set your project/site parameters in scripts/site_details first!"
  exit 1
fi

gcloud iam service-accounts create terraform --display-name "Terraform account"
gcloud iam service-accounts keys create terraform-key --iam-account terraform@"$TF_VAR_project_id".iam.gserviceaccount.com

export GOOGLE_APPLICATION_CREDENTIALS=./terraform-key

gcloud projects add-iam-policy-binding "$TF_VAR_project_id" \
  --member serviceAccount:terraform@"$TF_VAR_project_id".iam.gserviceaccount.com \
  --role roles/owner

gcloud projects add-iam-policy-binding "$TF_VAR_project_id" \
  --member serviceAccount:terraform@"$TF_VAR_project_id".iam.gserviceaccount.com \
  --role roles/storage.admin

terraform init -reconfigure

gcloud services enable cloudresourcemanager.googleapis.com
