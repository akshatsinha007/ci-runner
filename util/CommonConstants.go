/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import "path/filepath"

const (
	DEVTRON                      = "DEVTRON"
	DEFAULT_KEY                  = "default"
	INSECURE                     = "insecure"
	SECUREWITHCERT               = "secure-with-cert"
	DOCKER_PS_START_WAIT_SECONDS = 150
	HOMEDIR                      = "/"
	WORKINGDIR                   = "/devtroncd"
	TerminalLogDir               = "/dev"
	TerminalLogFile              = "/termination-log"
	LOCAL_BUILDX_LOCATION        = "/var/lib/devtron/buildx"
	LOCAL_BUILDX_CACHE_LOCATION  = LOCAL_BUILDX_LOCATION + "/cache"
	CIEVENT                      = "CI"
	JOBEVENT                     = "JOB"
	CDSTAGE                      = "CD"
	DRY_RUN                      = "DryRun"
	ENV_VARIABLE_BUILD_SUCCESS   = "BUILD_SUCCESS"
	CiCdEventEnvKey              = "CI_CD_EVENT"
	Source_Signal                = "Source_Signal"
	Source_Defer                 = "Source_Defer"
	DefaultErrorCode             = 1
	AbortErrorCode               = 143
	CiStageFailErrorCode         = 2
	InAppLogging                 = "IN_APP_LOGGING"
	CiRunnerCommand              = "./cirunner"
	TeeCommand                   = "tee"
	LogFileName                  = "main.log"
	NewLineChar                  = "\n"
	ArtifactSourceType           = "CI-RUNNER"
	ArtifactMaterialType         = "git"
	TerminationLogDir            = "/dev/termination-log"
)

const (
	ResultsDirInCIRunnerPath = "/polling-plugin/results.json"
	PluginArtifactsResults   = "/tmp/pluginArtifacts/results.json"
)

var (
	TmpArtifactLocation = "./job-artifact"
	TmpLogLocation      = "/main.log"
	Output_path         = filepath.Join(WORKINGDIR, "./process")

	Bash_script = filepath.Join("_script.sh")
)

type CiFailReason string

type CdFailReason string

func (r CiFailReason) String() string {
	return string(r)
}

func (r CdFailReason) String() string {
	return string(r)
}

const (
	PreCiFailed  CiFailReason = "Pre-CI task failed: %s"
	PostCiFailed CiFailReason = "Post-CI task failed: %s"
	BuildFailed  CiFailReason = "Docker build failed"
	PushFailed   CiFailReason = "Docker push failed"
	ScanFailed   CiFailReason = "Image scan failed"
	CiFailed     CiFailReason = "CI Failed: exit code 1"

	CdStageTaskFailed CdFailReason = "%s task failed: %s"
	CdStageFailed     CdFailReason = "%s failed"
)
