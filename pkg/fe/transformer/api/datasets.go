package api

import (
	"fmt"
	"lunchpail.io/pkg/fe/linker/queue"
	"lunchpail.io/pkg/ir/hlir"
	"lunchpail.io/pkg/util"
	"path/filepath"
)

type nfs struct {
	Path   string `json:"path"`
	Server string `json:"server"`
}

type pvc struct {
	ClaimName string `json:"claimName"`
}

type volume struct {
	Name                  string `json:"name"`
	Nfs                   *nfs   `json:"nfs,omitempty"`
	PersistentVolumeClaim *pvc   `json:"persistentVolumeClaim,omitempty"`
}

type volumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath,omitempty"`
}

type initContainer struct {
	Name         string        `json:"name"`
	Image        string        `json:"image"`
	Command      []string      `json:"command"`
	EnvFrom      []envFrom     `json:"envFrom"`
	VolumeMounts []volumeMount `json:"volumeMounts"`
}

type secretRef struct {
	Name string `json:"name"`
}

type envFrom struct {
	SecretRef secretRef `json:"secretRef"`
	Prefix    string    `json:"prefix,omitempty"`
}

func envForQueue(queueSpec queue.Spec) envFrom {
	return envFrom{secretRef{queueSpec.Name}, queueSpec.Name + "_"}
}

func datasets(app hlir.Application, queueSpec queue.Spec) ([]volume, []volumeMount, []envFrom, []initContainer, error) {
	volumes := []volume{}
	volumeMounts := []volumeMount{}
	envFroms := []envFrom{envForQueue(queueSpec)}
	initContainers := []initContainer{}

	for _, dataset := range app.Spec.Datasets {
		name := dataset.Name

		if dataset.Nfs.Server != "" {
			v := volume{}
			v.Name = name
			v.Nfs = &nfs{dataset.Nfs.Server, dataset.Nfs.Path}
			volumes = append(volumes, v)
			volumeMounts = append(volumeMounts, volumeMount{name, dataset.MountPath})
		}
		if dataset.Pvc.ClaimName != "" {
			v := volume{}
			v.Name = name
			v.PersistentVolumeClaim = &pvc{dataset.Pvc.ClaimName}
			volumes = append(volumes, v)
			volumeMounts = append(volumeMounts, volumeMount{name, dataset.MountPath})
		}
		if dataset.S3.Secret != "" {
			env := envFrom{secretRef{dataset.S3.Secret}, dataset.Name + "_"}
			// envFroms = append(envFroms, env)

			if dataset.S3.CopyIn.Path != "" {
				initContainers = append(initContainers, initContainer{
					Name:         "s3-copy-in-" + dataset.Name,
					Image:        "docker.io/rclone/rclone:1",
					EnvFrom:      []envFrom{env},
					VolumeMounts: []volumeMount{volumeMount{"workdir", "/workdir"}},
					Command: []string{
						"/bin/sh",
						"-c",
						fmt.Sprintf(`
printenv

config=/tmp/rclone.conf
cat <<EOF > $config
[s3]
type = s3
provider = Other
env_auth = false
endpoint = $%sendpoint
access_key_id = $%saccessKeyID
secret_access_key = $%ssecretAccessKey
acl = public-read
EOF

rclone --config $config config dump

rclone --config $config copyto s3:/%s /workdir/%s/%s
`, env.Prefix, env.Prefix, env.Prefix, dataset.S3.CopyIn.Path, dataset.Name, filepath.Base(dataset.S3.CopyIn.Path)),
					},
				})
			}
		}
	}

	return volumes, volumeMounts, envFroms, initContainers, nil
}

func datasetsB64(app hlir.Application, queueSpec queue.Spec) (string, string, string, string, error) {
	volumes, volumeMounts, envFroms, initContainers, err := datasets(app, queueSpec)
	if err != nil {
		return "", "", "", "", err
	}

	volumesB64, err := util.ToJsonB64(volumes)
	if err != nil {
		return "", "", "", "", err
	}

	volumeMountsB64, err := util.ToJsonB64(volumeMounts)
	if err != nil {
		return "", "", "", "", err
	}

	envFromsB64, err := util.ToJsonB64(envFroms)
	if err != nil {
		return "", "", "", "", err
	}

	initContainersB64, err := util.ToJsonB64(initContainers)
	if err != nil {
		return "", "", "", "", err
	}

	return volumesB64, volumeMountsB64, envFromsB64, initContainersB64, nil
}
