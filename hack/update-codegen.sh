#!/usr/bin/env bash

# Copyright 2023 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

# Get the code-generator version and root
CODEGEN_VERSION=$(go list -m -f '{{.Version}}' k8s.io/code-generator)
CODEGEN_ROOT=$(go env GOMODCACHE)/k8s.io/code-generator@${CODEGEN_VERSION}

# Use kube_codegen.sh directly from the module cache (no copying needed)
source "${CODEGEN_ROOT}/kube_codegen.sh"

THIS_PKG="github.com/xsts-sh/xstatefulset"

# Generate deepcopy, defaulter, conversion functions
kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    "${SCRIPT_ROOT}/api"

# Generate defaulter code explicitly
# This ensures SetObjectDefaults_* wrapper functions are generated
echo "Generating defaulter code..."
GOPROXY=off go install k8s.io/code-generator/cmd/defaulter-gen
# The result file of defaulter generation
DEFAULTER_OUTPUT_FILE="zz_generated.defaults.go"
# Find all directories that request defaulter generation
DEFAULTER_DIRS=$(find "${SCRIPT_ROOT}/api" -name "doc.go" -exec grep -l "+k8s:defaulter-gen=" {} \; | xargs -n1 dirname | sort -u)

if [ -n "${DEFAULTER_DIRS}" ]; then
    DEFAULTER_PKGS=""
    for dir in ${DEFAULTER_DIRS}; do
        # Convert absolute path to relative package path
        rel_dir=${dir#${SCRIPT_ROOT}/}
        DEFAULTER_PKGS="${DEFAULTER_PKGS} ./${rel_dir}"
    done
    echo "Running defaulter-gen for: ${DEFAULTER_PKGS}"
    # Remove old generated defaulter files
    find "${SCRIPT_ROOT}/api" -name "${DEFAULTER_OUTPUT_FILE}" -delete
    # Run defaulter-gen
    defaulter-gen \
        --go-header-file "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
        --output-file "${DEFAULTER_OUTPUT_FILE}" \
        ${DEFAULTER_PKGS}

    echo "Defaulter code generation completed"
else
    echo "No directories found with +k8s:defaulter-gen tags"
fi

# Define external apply configurations for core/v1 types
EXTERNAL_APPLYCONFIGS="k8s.io/api/core/v1.PodTemplateSpec:k8s.io/client-go/applyconfigurations/core/v1"
EXTERNAL_APPLYCONFIGS="${EXTERNAL_APPLYCONFIGS},k8s.io/api/core/v1.PersistentVolumeClaim:k8s.io/client-go/applyconfigurations/core/v1"
EXTERNAL_APPLYCONFIGS="${EXTERNAL_APPLYCONFIGS},k8s.io/api/autoscaling/v1.Scale:k8s.io/client-go/applyconfigurations/autoscaling/v1"
EXTERNAL_APPLYCONFIGS="${EXTERNAL_APPLYCONFIGS},k8s.io/api/apps/v1.PodManagementPolicyType:k8s.io/client-go/applyconfigurations/apps/v1"
EXTERNAL_APPLYCONFIGS="${EXTERNAL_APPLYCONFIGS},k8s.io/api/apps/v1.StatefulSetUpdateStrategy:k8s.io/client-go/applyconfigurations/apps/v1"
EXTERNAL_APPLYCONFIGS="${EXTERNAL_APPLYCONFIGS},k8s.io/api/apps/v1.StatefulSetPersistentVolumeClaimRetentionPolicy:k8s.io/client-go/applyconfigurations/apps/v1"
EXTERNAL_APPLYCONFIGS="${EXTERNAL_APPLYCONFIGS},k8s.io/api/apps/v1.StatefulSetOrdinals:k8s.io/client-go/applyconfigurations/apps/v1"
EXTERNAL_APPLYCONFIGS="${EXTERNAL_APPLYCONFIGS},k8s.io/api/apps/v1.StatefulSetCondition:k8s.io/client-go/applyconfigurations/apps/v1"

# Generate client, listers, informers and apply configurations
kube::codegen::gen_client \
    --with-watch \
    --with-applyconfig \
    --applyconfig-externals "${EXTERNAL_APPLYCONFIGS}" \
    --output-dir "${SCRIPT_ROOT}/client-go" \
    --output-pkg "${THIS_PKG}/client-go" \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    "${SCRIPT_ROOT}/api"

# Fix the autoscaling/v1 import in generated client code
# client-gen doesn't know that autoscaling/v1 apply configurations come from k8s.io/client-go
find "${SCRIPT_ROOT}/client-go/clientset" -name "*.go" -type f -exec sed -i \
    -e 's|applyconfigurationautoscalingv1 "github.com/xsts-sh/xstatefulset/client-go/applyconfiguration/autoscaling/v1"|applyconfigurationautoscalingv1 "k8s.io/client-go/applyconfigurations/autoscaling/v1"|g' \
    {} +


bash "${SCRIPT_ROOT}/hack/update-crd.sh"
