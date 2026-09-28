package queries

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestHasTTYDPort(t *testing.T) {
	tests := []struct {
		name  string
		ports []corev1.ServicePort
		want  bool
	}{
		{
			name: "stable ttyd discovery contract",
			ports: []corev1.ServicePort{{
				Name: "ttyd",
				Port: 7681,
			}},
			want: true,
		},
		{
			name: "generic clabernetes port name is rejected",
			ports: []corev1.ServicePort{{
				Name: "port-7681-tcp",
				Port: 7681,
			}},
			want: false,
		},
		{
			name: "ttyd name on another port is rejected",
			ports: []corev1.ServicePort{{
				Name: "ttyd",
				Port: 8080,
			}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasTTYDPort(tt.ports); got != tt.want {
				t.Fatalf("hasTTYDPort() = %v, want %v", got, tt.want)
			}
		})
	}
}
