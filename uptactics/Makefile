init-workspaces:
	terraform workspace new staging
	terraform workspace new production

plan-staging:
	terraform workspace select staging
	terraform plan -var-file="staging.tfvars"

plan-production:
	terraform workspace select production
	terraform plan -var-file="production.tfvars"

apply-staging:
	terraform workspace select staging
	terraform apply -var-file="staging.tfvars"

apply-production:
	terraform workspace select production
	terraform apply -var-file="production.tfvars"
