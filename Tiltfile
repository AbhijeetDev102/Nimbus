k8s_yaml("./infra/development/k8s/secret.yml")
k8s_yaml("./infra/development/k8s/configmap.yml")
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
    cmd="curl --retry 15 --retry-all-errors --retry-connrefused --retry-delay 2 -s -X POST -H 'Content-Type: application/json' -d @infra/development/k8s/register-postgres-outbox.json http://localhost:8083/connectors",
    resource_deps=["connect"],
)

# ----------------------------------------------------
# Microservices Container Builds & Deployments
# ----------------------------------------------------

docker_build("nimbus-api-gateway", ".", dockerfile="./services/api-gateway/Dockerfile")
docker_build("nimbus-resource-service", ".", dockerfile="./services/resource-service/Dockerfile")
docker_build("nimbus-job-service", ".", dockerfile="./services/job-service/Dockerfile")
docker_build("nimbus-worker-service", ".", dockerfile="./services/worker-service/Dockerfile")

k8s_yaml("./infra/development/k8s/api-gateway-deployment.yml")
k8s_yaml("./infra/development/k8s/resource-service-deployment.yml")
k8s_yaml("./infra/development/k8s/job-service-deployment.yml")
k8s_yaml("./infra/development/k8s/worker-service-deployment.yml")

k8s_resource("api-gateway", port_forwards='8081:8081')
k8s_resource("resource-service", resource_deps=["postgres", "minio"])
k8s_resource("job-service", resource_deps=["postgres"])
k8s_resource("worker-service", resource_deps=["postgres", "minio", "kafka"])


k8s_yaml("./infra/development/k8s/redis-deployment.yml")
k8s_resource("redis", port_forwards='6379:6379')

# ----------------------------------------------------
# Nimbus Developer Studio & Control Plane Dashboard (Vite + React)
# ----------------------------------------------------
docker_build(
    "nimbus-dashboard",
    "./web",
    dockerfile="./web/Dockerfile",
    ignore=["./web/node_modules", "./web/dist"]
)

k8s_yaml("./infra/development/k8s/dashboard-deployment.yml")
k8s_resource("nimbus-dashboard", port_forwards='3000:3000', links=['http://localhost:3000'], resource_deps=["api-gateway"])


