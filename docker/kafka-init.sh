kafka-topics --bootstrap-server broker:29092 --create --if-not-exists --topic dev.reporting.sale --replication-factor 1 --partitions 12
echo "created dev.reporting.sale topic"

kafka-topics --bootstrap-server broker:29092 --create --if-not-exists --topic qa.reporting.sale --replication-factor 1 --partitions 12
echo "created qa.reporting.sale topic"

kafka-topics --bootstrap-server broker:29092 --create --if-not-exists --topic dev.reporting.country --replication-factor 1 --partitions 12
echo "created dev.reporting.country topic"

kafka-topics --bootstrap-server broker:29092 --create --if-not-exists --topic qa.reporting.country --replication-factor 1 --partitions 12
echo "created qa.reporting.country topic"

for i in {1..50}; do
  for env in "dev" "qa"; do
    echo "{\"id\": $i}" | kafka-console-producer --bootstrap-server broker:29092 --topic $env.reporting.sale &
    echo "Produced message $i to topic $env.reporting.sale"
  done
done

for env in "dev" "qa"; do
  for i in {1..2}; do
    kafka-console-consumer --bootstrap-server broker:29092 --topic $env.reporting.sale --group ktea-test-group-$i --from-beginning --timeout-ms 5000
  done
done