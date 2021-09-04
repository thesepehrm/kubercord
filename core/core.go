package core

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thesepehrm/kubercord/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	errNoMatches    = errors.New("No matches")
	importantPhases = []string{"FailedScheduling", "FailedSync", "FailedValidation", "Failed", "Succeeded", "Running", "Completed", "Pending"}
)

type KuberCord struct {
	clientset   *kubernetes.Clientset
	dh          *DiscordHandler
	podsWatcher watch.Interface
	podsCache   map[string]*v1.Pod
	q           chan struct{}
}

func Init() (*KuberCord, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", config.Config.K8s.ConfigDir)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	dh, err := NewDiscordHandler()
	if err != nil {
		return nil, err
	}

	return &KuberCord{
		clientset: clientset,
		dh:        dh,
		q:         make(chan struct{}),
		podsCache: make(map[string]*v1.Pod),
	}, nil
}

func (kc *KuberCord) Start() {
	var err error
	kc.podsWatcher, err = kc.clientset.CoreV1().Pods(config.Config.K8s.Namespace).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		logrus.WithError(err).Panic("Failed to watch pods")
	}
	logrus.Info("Watching pods")
	go kc.Watch()
	go kc.watchForEvents()
}

func (kc *KuberCord) Stop() {
	kc.q <- struct{}{}
	close(kc.q)
}

func (kc *KuberCord) Watch() {
	for {
		select {
		case <-kc.q:
			kc.podsWatcher.Stop()
			return
		default:
			kc.watch()
		}
		time.Sleep(time.Duration(config.Config.K8s.Interval) * time.Second)
	}
}

func (kc *KuberCord) watchForEvents() {
	for event := range kc.podsWatcher.ResultChan() {
		if event.Object.GetObjectKind().GroupVersionKind().Kind == "Pod" {
			pod := event.Object.(*v1.Pod)
			kc.handlePod(pod)
		}
	}
}

func (kc *KuberCord) watch() {
	n := config.Config.K8s.Namespace
	s := int64(config.Config.K8s.Interval)
	pods, err := kc.clientset.CoreV1().Pods(config.Config.K8s.Namespace).List(context.Background(), metav1.ListOptions{})

	if err != nil {
		logrus.WithError(err).Panic("Failed to list pods")
	}
	logrus.WithField("pods-count", len(pods.Items)).Info("Number of pods")

	for _, pod := range pods.Items {
		resp := kc.clientset.CoreV1().Pods(n).GetLogs(pod.Name, &v1.PodLogOptions{
			SinceSeconds: &s,
		}).Do(context.Background())

		rawLogs, err := resp.Raw()
		if err != nil {
			logrus.WithError(err).Panic("Failed to get logs")
		}
		logs := string(rawLogs)
		a := parseLogs(pod.Name, logs)
		if a != nil {
			err := kc.dh.SendAlert(a)
			if err != nil {
				logrus.WithFields(
					logrus.Fields{
						"service": a.service,
						"error":   err,
					}).Error("Failed to send alert")
			}
		}
	}
}

func parseLogs(service, logs string) *Alert {
	splitted := strings.Split(logs, "\r")
	for _, line := range splitted {
		params, err := extractParams(line)

		if err != nil {
			if err == errNoMatches {
				continue
			}

			logrus.WithFields(logrus.Fields{
				"service": service,
				"error":   err,
			}).Error("Failed to parse logs")
			continue
		}
		if params.Level == "fatal" {
			return NewAlert(service, Fatal, params.Msg, params.Time, splitted)
		}
		if params.Level == "error" {
			return NewAlert(service, Error, params.Msg, params.Time, splitted)
		}
		/*if params.Level == "warn" {
			return NewAlert(service, Warn, params.Msg, params.Time, splitted)
		}*/
		/*if params.Level == "info" {
			return NewAlert(service, Info, params.Msg, params.Time, splitted)
		}
		if params.Level == "debug" {
			return NewAlert(service, Debug, params.Msg, params.Time, splitted)
		}*/
	}
	return nil
}

func (kc *KuberCord) handlePod(newPod *v1.Pod) {
	if p, ok := kc.podsCache[newPod.Name]; ok {
		if p.Status.Phase == newPod.Status.Phase {
			return
		}
	}

	kc.podsCache[newPod.Name] = newPod

	if arrayContains(importantPhases, string(newPod.Status.Phase)) {
		logrus.WithFields(logrus.Fields{
			"pod":    newPod.Name,
			"status": newPod.Status.Phase,
		}).Info("Pod status changed")
		err := kc.dh.SendPodStatus(newPod.Name, string(newPod.Status.Phase), newPod.Status.Message)
		if err != nil {
			logrus.WithError(err).Error("Failed to send pod status")
		}
	}

}
