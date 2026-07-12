k8s_yaml("./infra/development/k8s/secret.yml")
k8s_yaml("./infra/development/k8s/minio-deployment.yml")
k8s_yaml("./infra/development/k8s/postgres-deployment.yml")

k8s_resource("minio", port_forwards=['9000:9000','9001:9001'])

k8s_resource("postgres", port_forwards='5432:5432')