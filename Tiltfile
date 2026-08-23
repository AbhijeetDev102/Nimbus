k8s_yaml("./infra/development/k8s/secret.yml")
k8s_yaml("./infra/development/k8s/minio-deployment.yml")
k8s_yaml("./infra/development/k8s/postgres-deployment.yml")

k8s_resource("minio", port_forwards=['9000:9000','9001:9001'])

k8s_resource("postgres", port_forwards='5432:5432')

k8s_yaml("./infra/development/k8s/kafka-deployment.yml")
k8s_yaml("./infra/development/k8s/debezium-kafka-connect-deployment.yml")

k8s_resource("kafka", port_forwards='9092:9092')
k8s_resource("connect", port_forwards='8083:8083')

local_resource(
    name="register-outbox-connector",
    cmd="go run ./tools/registerConnector/register_connector.go",
    resource_deps=["connect"],
)
