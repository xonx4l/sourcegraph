"""
Load external dependencies for base images
"""

load("@rules_oci//oci:pull.bzl", "oci_pull")

# Quick script to get the latest tags for each of the base images from GCR:
#
# grep 'image = ' ./dev/oci_deps.bzl | while read -r str ; do
#   str_no_spaces="${str#"${str%%[![:space:]]*}"}"  # remove leading spaces
#   url="${str_no_spaces#*\"}"  # remove prefix until first quote
#   url="${url%%\"*}"  # remove suffix from first quote
#
#   IMAGE_DETAILS=$(gcloud container images list-tags $url --limit=1 --sort-by=~timestamp --format=json)
#   TAG=$(echo $IMAGE_DETAILS | jq -r '.[0].tags[0]')
#   DIGEST=$(echo $IMAGE_DETAILS | jq -r '.[0].digest')
#
#   echo $url
#   echo $DIGEST
# done
#
#
# Quick script to get the latest tags for each of the base images from Dockerhub:
# grep 'image = ' ./dev/oci_deps.bzl | while read -r str ; do
#   str_no_spaces="${str#"${str%%[![:space:]]*}"}"  # remove leading spaces
#   url="${str_no_spaces#*\"}"  # remove prefix until first quote
#   url="${url%%\"*}"  # remove suffix from first quote

#     TOKEN=$(curl -s "https://auth.docker.io/token?service=registry.docker.io&scope=repository:${url}:pull" | jq -r .token)

#   DIGEST=$(curl -I -s -H "Authorization: Bearer $TOKEN" -H "Accept: application/vnd.docker.distribution.manifest.v2+json" \
#     https://registry-1.docker.io/v2/${url}/manifests/latest \
#     | grep -i Docker-Content-Digest | awk '{print $2}')

#   echo -e "$url\n$DIGEST\n\n"
# done

def oci_deps():
    """
    The image definitions and their digests
    """
    oci_pull(
        name = "wolfi_base",
        digest = "sha256:1c9377d33e1c127d8949b1d6b6b9109e51decfe3184d69485a0c8fe96a7bb999",
        image = "index.docker.io/sourcegraph/wolfi-sourcegraph-base",
    )

    oci_pull(
        name = "wolfi_cadvisor_base",
        digest = "sha256:ffc31501d24dcc70e0d1527f11be8eb92812b4fc6fdf69f6b134dee53aa7c302",
        image = "index.docker.io/sourcegraph/wolfi-cadvisor-base",
    )

    oci_pull(
        name = "wolfi_symbols_base",
        digest = "sha256:8f2976b67f4488a015cce458108868dc15a6864a25f9ad892b6d0390abe0305a",
        image = "index.docker.io/sourcegraph/wolfi-symbols-base",
    )

    oci_pull(
        name = "wolfi_server_base",
        digest = "sha256:3edcdef4f679c5495e6b617858fc28788c18a026bd1ab6bdbcd847ff1d0b125b",
        image = "index.docker.io/sourcegraph/wolfi-server-base",
    )

    oci_pull(
        name = "wolfi_gitserver_base",
        digest = "sha256:12eeb43a2e443ad4afcf9a9b68db308e3af57c6399383d0f491bb2a8fc2bca20",
        image = "index.docker.io/sourcegraph/wolfi-gitserver-base",
    )

    oci_pull(
        name = "wolfi_grafana_base",
        digest = "sha256:461ef0c0233c20716398a0ad1a38dc2df955e26ec091400f64a18fa7de9d2621",
        image = "us.gcr.io/sourcegraph-dev/wolfi-grafana",
    )

    oci_pull(
        name = "wolfi_postgres_exporter_base",
        digest = "sha256:ad22636ac73eb77fdf733f4a77d359988bb7ad59267db3762cd60f0741ba2ecb",
        image = "index.docker.io/sourcegraph/wolfi-postgres-exporter-base",
    )

    oci_pull(
        name = "wolfi_jaeger_all_in_one_base",
        digest = "sha256:62134cf565429c76279f6aa6fc25f8ade4168ab6e252ea642b4f12914c1b615a",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-all-in-one-base",
    )

    oci_pull(
        name = "wolfi_jaeger_agent_base",
        digest = "sha256:958793cb5e6b963834e9f25cb0e0c1765ae13e6c074dd9733e49ee8d4c1d2b5e",
        image = "index.docker.io/sourcegraph/wolfi-jaeger-agent-base",
    )

    oci_pull(
        name = "wolfi_redis_base",
        digest = "sha256:e415bcffc676c2fd641dfd54d48af67314899f6c67dc1989980d16507907c679",
        image = "index.docker.io/sourcegraph/wolfi-redis-base",
    )

    oci_pull(
        name = "wolfi_redis_exporter_base",
        digest = "sha256:51780ffd7c142f992d4b56b7aee08f54add8acd5fcffe65f0bc5115353f60985",
        image = "index.docker.io/sourcegraph/wolfi-redis-exporter-base",
    )

    oci_pull(
        name = "wolfi_syntax_highlighter_base",
        digest = "sha256:1af5c0050ef10256802a71906e3775710af1be14dadcbad9da3c6f30e659934c",
        image = "index.docker.io/sourcegraph/wolfi-syntax-highlighter-base",
    )

    oci_pull(
        name = "wolfi_search_indexer_base",
        digest = "sha256:215b187278142691880b87979b89bb83617a4f56a0990e69aeb12f2ddbb96a0a",
        image = "index.docker.io/sourcegraph/wolfi-search-indexer-base",
    )

    oci_pull(
        name = "wolfi_repo_updater_base",
        digest = "sha256:e674cfc018f5872a4b45ce567071f1587f6f4fb0df79f299ad135c1bb1d39c0c",
        image = "index.docker.io/sourcegraph/wolfi-repo-updater-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:bcb941e88f7c02010aabcab08dd5e7924a3523e2c5fa0e579c6b462e0bedc473",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_executor_base",
        digest = "sha256:f0e902355468f35049d5233c36c154870d757429e07888fa5096a7d09bd0bc29",
        image = "index.docker.io/sourcegraph/wolfi-executor-base",
    )

    # ???
    oci_pull(
        name = "wolfi_bundled_executor_base",
        digest = "sha256:a1d8a4875e5f6760774b9cd69dc28fb8760c4b5d09736506c711dbbd84df168d",
        image = "index.docker.io/sourcegraph/wolfi-bundled-executor-base",
    )

    oci_pull(
        name = "wolfi_executor_kubernetes_base",
        digest = "sha256:6e351096846afd4b56dc207805746d7b7b37fdcf217928cc1fec16fa10abc7bf",
        image = "index.docker.io/sourcegraph/wolfi-executor-kubernetes-base",
    )

    oci_pull(
        name = "wolfi_batcheshelper_base",
        digest = "sha256:abd80b65ac08713d8ea3fa24dd64e80ea1bcf4a01ed327fa90ef834c63025753",
        image = "index.docker.io/sourcegraph/wolfi-batcheshelper-base",
    )

    oci_pull(
        name = "wolfi_prometheus_base",
        digest = "sha256:d6f84a0a85c2a2e91168d642f3b5c82d8f215d9a017b3d4bd6920cadb198b7b0",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-base",
    )

    oci_pull(
        name = "wolfi_prometheus_gcp_base",
        digest = "sha256:dcd077ea58b13624f48ad119cbca21b7e9fe95082178514edccec4f677b148f5",
        image = "index.docker.io/sourcegraph/wolfi-prometheus-gcp-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12_base",
        digest = "sha256:78f803bfecef59f870fe4962107d5301bc05855d77cf21db86cec9ed04ec8075",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-base",
    )

    oci_pull(
        name = "wolfi_postgresql-12-codeinsights_base",
        digest = "sha256:4c34dbf0d9e51fbbbc26a43b8ff4503eb67382082b98fc27f9f618a2254a6bd3",
        image = "index.docker.io/sourcegraph/wolfi-postgresql-12-codeinsights-base",
    )

    oci_pull(
        name = "wolfi_node_exporter_base",
        digest = "sha256:83241641f50a02bb19b0dd962e3c9bb8d863fa3bd46f39126c42156a11b592f1",
        image = "index.docker.io/sourcegraph/wolfi-node-exporter-base",
    )

    oci_pull(
        name = "wolfi_opentelemetry_collector_base",
        digest = "sha256:3f99e520475c71e07ce2b35b21d2f26a9892f4058810b4f1e61a57251eb84406",
        image = "index.docker.io/sourcegraph/wolfi-opentelemetry-collector-base",
    )

    oci_pull(
        name = "wolfi_searcher_base",
        digest = "sha256:bcb941e88f7c02010aabcab08dd5e7924a3523e2c5fa0e579c6b462e0bedc473",
        image = "index.docker.io/sourcegraph/wolfi-searcher-base",
    )

    oci_pull(
        name = "wolfi_s3proxy_base",
        digest = "sha256:fdf26ade72fd30a557f635a14550c4583b8e8427f857302466bc2c7630578e2f",
        image = "index.docker.io/sourcegraph/wolfi-blobstore-base",
    )

    oci_pull(
        name = "wolfi_qdrant_base",
        digest = "sha256:630d94560574ac94ad03f9bb310c8dcdff8c5a7f9a250336426f69077549c561",
        image = "index.docker.io/sourcegraph/wolfi-qdrant-base",
    )
