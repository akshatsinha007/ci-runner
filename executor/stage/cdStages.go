/*
 *  Copyright 2020 Devtron Labs
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
 *
 */

package stage

import (
	"github.com/devtron-labs/ci-runner/executor"
	util2 "github.com/devtron-labs/ci-runner/executor/util"
	"github.com/devtron-labs/ci-runner/helper"
	"github.com/devtron-labs/ci-runner/util"
	"log"
	"os"
)

type CdStage struct {
	GitManager helper.GitManager
}

func NewCdStage(GitManager helper.GitManager) *CdStage {
	return &CdStage{
		GitManager: GitManager,
	}
}

func (impl CdStage) HandleCDEvent(ciCdRequest *helper.CiCdTriggerEvent, exitCode *int) {
	err := impl.runCDStages(ciCdRequest)
	artifactUploadErr := collectAndUploadCDArtifacts(ciCdRequest.CommonWorkflowRequest)
	if err != nil || artifactUploadErr != nil {
		log.Println(err)
		*exitCode = util.DefaultErrorCode
	}

}

func collectAndUploadCDArtifacts(cdRequest *helper.CommonWorkflowRequest) error {
	cloudHelperBaseConfig := cdRequest.GetCloudHelperBaseConfig(util.BlobStorageObjectTypeArtifact)
	if cdRequest.PrePostDeploySteps != nil && len(cdRequest.PrePostDeploySteps) > 0 {
		_, err := helper.ZipAndUpload(cloudHelperBaseConfig, cdRequest.CiArtifactFileName)
		return err
	}

	// to support stage YAML outputs
	artifactFiles := make(map[string]string)
	var allTasks []*helper.Task
	if cdRequest.TaskYaml != nil {
		for _, pc := range cdRequest.TaskYaml.CdPipelineConfig {
			for _, t := range append(pc.BeforeTasks, pc.AfterTasks...) {
				allTasks = append(allTasks, t)
			}
		}
	}
	for _, task := range allTasks {
		if task.RunStatus {
			if _, err := os.Stat(task.OutputLocation); os.IsNotExist(err) { // Ignore if no file/folder
				log.Println(util.DEVTRON, "artifact not found ", err)
				continue
			}
			artifactFiles[task.Name] = task.OutputLocation
		}
	}
	log.Println(util.DEVTRON, " artifacts", artifactFiles)
	return helper.UploadArtifact(cloudHelperBaseConfig, artifactFiles, cdRequest.CiArtifactFileName)
}

func (impl CdStage) runCDStages(cicdRequest *helper.CiCdTriggerEvent) error {
	err := os.Chdir("/")
	if err != nil {
		return err
	}

	if _, err := os.Stat(util.WORKINGDIR); os.IsNotExist(err) {
		_ = os.Mkdir(util.WORKINGDIR, os.ModeDir)
	}
	err = os.Chdir(util.WORKINGDIR)
	if err != nil {
		return err
	}
	// git handling
	// we are skipping clone and checkout in case of ci job type poll cr images plugin does not require it.(ci-job)
	skipCheckout := cicdRequest.CommonWorkflowRequest.CiPipelineType == helper.CI_JOB
	if !skipCheckout {
		log.Println(util.DEVTRON, " git")
		err = impl.GitManager.CloneAndCheckout(cicdRequest.CommonWorkflowRequest.CiProjectDetails)
		if err != nil {
			log.Println(util.DEVTRON, "clone err: ", err)
			return err
		}
	}
	log.Println(util.DEVTRON, " /git")
	// Start docker daemon
	log.Println(util.DEVTRON, " docker-start")
	helper.StartDockerDaemon(cicdRequest.CommonWorkflowRequest.DockerConnection, cicdRequest.CommonWorkflowRequest.DockerRegistryURL, cicdRequest.CommonWorkflowRequest.DockerCert, cicdRequest.CommonWorkflowRequest.DefaultAddressPoolBaseCidr, cicdRequest.CommonWorkflowRequest.DefaultAddressPoolSize, -1)

	err = helper.DockerLogin(&helper.DockerCredentials{
		DockerUsername:     cicdRequest.CommonWorkflowRequest.DockerUsername,
		DockerPassword:     cicdRequest.CommonWorkflowRequest.DockerPassword,
		AwsRegion:          cicdRequest.CommonWorkflowRequest.AwsRegion,
		AccessKey:          cicdRequest.CommonWorkflowRequest.AccessKey,
		SecretKey:          cicdRequest.CommonWorkflowRequest.SecretKey,
		DockerRegistryURL:  cicdRequest.CommonWorkflowRequest.DockerRegistryURL,
		DockerRegistryType: cicdRequest.CommonWorkflowRequest.DockerRegistryType,
	})
	if err != nil {
		return err
	}

	scriptEnvs, err := util2.GetGlobalEnvVariables(cicdRequest)

	if len(cicdRequest.CommonWorkflowRequest.PrePostDeploySteps) > 0 {
		refStageMap := make(map[int][]*helper.StepObject)
		for _, ref := range cicdRequest.CommonWorkflowRequest.RefPlugins {
			refStageMap[ref.Id] = ref.Steps
		}
		scriptEnvs["DEST"] = cicdRequest.CommonWorkflowRequest.CiArtifactDTO.Image
		scriptEnvs["DIGEST"] = cicdRequest.CommonWorkflowRequest.CiArtifactDTO.ImageDigest
		var stage = executor.StepType(cicdRequest.CommonWorkflowRequest.StageType)
		_, _, err = executor.RunCiCdSteps(stage, cicdRequest.CommonWorkflowRequest.PrePostDeploySteps, refStageMap, scriptEnvs, nil)
		if err != nil {
			return err
		}
	} else {

		// Get devtron-cd yaml
		taskYaml, err := helper.ToTaskYaml([]byte(cicdRequest.CommonWorkflowRequest.StageYaml))
		if err != nil {
			log.Println(err)
			return err
		}
		cicdRequest.CommonWorkflowRequest.TaskYaml = taskYaml

		// run post artifact processing
		log.Println(util.DEVTRON, " stage yaml", taskYaml)
		var tasks []*helper.Task
		for _, t := range taskYaml.CdPipelineConfig {
			tasks = append(tasks, t.BeforeTasks...)
			tasks = append(tasks, t.AfterTasks...)
		}

		if err != nil {
			return err
		}
		err = executor.RunCdStageTasks(tasks, scriptEnvs)
		if err != nil {
			return err
		}
	}

	// dry run flag indicates that ci runner image is being run from external helm chart
	if !cicdRequest.CommonWorkflowRequest.IsDryRun {
		log.Println(util.DEVTRON, " event")
		err = helper.SendCDEvent(cicdRequest.CommonWorkflowRequest)
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println(util.DEVTRON, " /event")
	}
	err = helper.StopDocker()
	if err != nil {
		log.Println("error while stopping docker", err)
		return err
	}
	return nil
}
