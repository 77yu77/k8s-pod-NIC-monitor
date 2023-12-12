# k8s-pod-NIC-monitor
There are a prometheus podmonitor and an exporter to collect the NIC message for pods which add the specific label.
## Environment configuration
These components need to be run on a kubernetes cluster with prometheus installed. To install the prometheus in kubernetes cluster, please see https://github.com/prometheus/prometheus.
## podmonitor installation
To enable the podmonitor, the first step is to add the podmonitor match label. The config context is in the `prometheus-prometheus.yaml`, which defines the deployment of the Prometheus server and its configuration.
```
spec:
  podMonitorSelector:
    matchLabels:
      team: pod-NIC # your match label
```
And in the `podmonitor.yaml`, you should add the label for prometheus to match.
```
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: pod-NIC-exporter # promtheus jobs name
  namespace: monitoring
  labels:
    team: pod-NIC # your match label
spec:
  namespaceSelector:
    any: true    
  podMetricsEndpoints:
  - interval: 5s
    scrapeTimeout: 5s  
    path: /metrics
    targetPort: 2112
    port: prometheus # pod port name
  selector:
    matchLabels:
      k8s-app: pod-NIC # label for the collected pods
```
After rewriting `podmonitor.yaml`, you should apply this yaml:
```
kubectl apply -f podmonitor.yaml
```
## Add exporter to pod
The exporter is inserted into pod by using mount volume. The following context provides one way to mount. 
First you should create kubernetes pv and pvc for the `pod_NIC_exporter` file. The cluster should deploy storage plugin such as NFS.
If you create pod by yaml file, the following code you can refer:
```
metadata:
  labels:
    k8s-app: pod-NIC # label for the collected pods
 spec:
    containers:
    - ...
      ports:
        - containerPort: 2112
          name: prometheus # pod port name
      volumeMounts:
      - name: pod_NIC
        mountPath: /usr/share/NIC/pod_NIC_exporter                    
  volumes:
      - name: pod_NIC
        persistentVolumeClaim:
          claimName: pod_NIC_pvc # your pvc name 
```
If you create pod by using client-go, the following code you can refer:
```
pod.WithLabels(map[string]string{
			"k8s-app": "pod-NIC",
})
Containers: []v1.ContainerApplyConfiguration{
	{
  ...
		Ports: []v1.ContainerPortApplyConfiguration{
			{
				Name:          "prometheus", // pod port name
				ContainerPort: 2112,
			},
		},
		VolumeMounts: []v1.VolumeMountApplyConfiguration{
			{
				MountPath: "/usr/share/NIC/pod_NIC_exporter",
				Name:      "pod_NIC",
			},
		},
	},
},
Volumes: []v1.VolumeApplyConfiguration{
	{
		Name: "pod_NIC",
		VolumeSourceApplyConfiguration: v1.VolumeSourceApplyConfiguration{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSourceApplyConfiguration{
				ClaimName: "pod_NIC_pvc", // your pvc name
			},
		},
	},
}
```
Both two methods require you to add start cmd in the pod start cmd. A easy start way is to add this following cmd to the container start cmd:
```
./usr/share/NIC/pod_NIC_exporter &
```
After that you can watch the data in the prometheus query page.