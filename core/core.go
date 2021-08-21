package core

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thesepehrm/kubercord/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	errNoMatches = errors.New("No matches")
)

type KuberCord struct {
	clientset *kubernetes.Clientset
	dh        *DiscordHandler
	q         chan struct{}
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
	}, nil
}

func (kc *KuberCord) Start() {
	go kc.Watch()
}

func (kc *KuberCord) Stop() {
	kc.q <- struct{}{}
	close(kc.q)
}

func (kc *KuberCord) Watch() {
	for {
		select {
		case <-kc.q:
			return
		default:
			kc.watch()
		}
		time.Sleep(time.Duration(config.Config.K8s.Interval) * time.Second)
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
		if string(pod.Status.Phase) != "Running" && string(pod.Status.Phase) != "Completed" {
			logrus.WithFields(logrus.Fields{
				"pod":    pod.Name,
				"status": pod.Status.Phase,
				"reason": pod.Status.Reason,
			}).Error("Pod failed")
			kc.dh.SendPodStatus(pod.Name, string(pod.Status.Phase), pod.Status.Reason)
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
		if params.Level == "warn" {
			return NewAlert(service, Warn, params.Msg, params.Time, splitted)
		}
		/*if params.Level == "info" {
			return NewAlert(service, Info, params.Msg, params.Time, splitted)
		}
		if params.Level == "debug" {
			return NewAlert(service, Debug, params.Msg, params.Time, splitted)
		}*/
	}
	return nil
}

type Params struct {
	Time  time.Time
	Level string
	Msg   string
}

func extractParams(line string) (*Params, error) {
	r := regexp.MustCompile(`time="(.*)"\s+level=(.*)\s+msg="(.*)"`)
	matches := r.FindStringSubmatch(line)
	if len(matches) == 0 {
		return nil, errNoMatches
	}

	t, err := time.Parse("2006-01-02T15:04:05Z", matches[1])
	if err != nil {
		return nil, err
	}

	return &Params{
		Time:  t,
		Level: matches[2],
		Msg:   matches[3],
	}, nil

}
