package pkg

import (
	"fmt"

	"github.com/myfintech/ark/src/go/lib/kube/objects"
	corev1 "k8s.io/api/core/v1"
)

func buildInstallInfoConfigMap(opts *DatadogOptions) *corev1.ConfigMap {
	return objects.ConfigMap(objects.ConfigMapOptions{
		Name: "datadog-agent-installinfo",
		Data: map[string]string{
			"install_info": fmt.Sprintf(`---
install_method:
  tool: MANTL Ark
  tool_version: %s
  installer_version: %s`, opts.ArkVersion, opts.ArkVersion),
		},
	})
}

func buildSystemProbeConfigMap() *corev1.ConfigMap {
	return objects.ConfigMap(objects.ConfigMapOptions{
		Name:   "datadog-agent-system-probe-config",
		Labels: nil,
		Data: map[string]string{
			"system-probe.yaml": `system_probe_config:
  enabled: true
  debug_port:  0
  sysprobe_socket: /var/run/sysprobe/sysprobe.sock
  enable_conntrack: true
  bpf_debug: false
  enable_tcp_queue_length: false
  enable_oom_kill: false
  collect_dns_stats: true
runtime_security_config:
  enabled: false
  debug: false
  socket: /var/run/sysprobe/runtime-security.sock
  policies:
    dir: /etc/datadog-agent/runtime-security.d
  syscall_monitor:
    enabled: false`,
		},
	})
}

func buildSecurityConfigMap() *corev1.ConfigMap {
	return objects.ConfigMap(objects.ConfigMapOptions{
		Name: "datadog-agent-security",
		Data: map[string]string{
			"system-probe-seccomp.json": `{
  "defaultAction": "SCMP_ACT_ERRNO",
  "syscalls": [
    {
      "names": [
        "accept4",
        "access",
        "arch_prctl",
        "bind",
        "bpf",
        "brk",
        "capget",
        "capset",
        "chdir",
        "clock_gettime",
        "clone",
        "close",
        "connect",
        "copy_file_range",
        "creat",
        "dup",
        "dup2",
        "dup3",
        "epoll_create",
        "epoll_create1",
        "epoll_ctl",
        "epoll_ctl_old",
        "epoll_pwait",
        "epoll_wait",
        "epoll_wait",
        "epoll_wait_old",
        "eventfd",
        "eventfd2",
        "execve",
        "execveat",
        "exit",
        "exit_group",
        "fchmod",
        "fchmodat",
        "fchown",
        "fchown32",
        "fchownat",
        "fcntl",
        "fcntl64",
        "fstat",
        "fstat64",
        "fstatfs",
        "fsync",
        "futex",
        "getcwd",
        "getdents",
        "getdents64",
        "getegid",
        "geteuid",
        "getgid",
        "getpeername",
        "getpid",
        "getppid",
        "getpriority",
        "getrandom",
        "getresgid",
        "getresgid32",
        "getresuid",
        "getresuid32",
        "getrlimit",
        "getrusage",
        "getsid",
        "getsockname",
        "getsockopt",
        "gettid",
        "gettimeofday",
        "getuid",
        "getxattr",
        "ioctl",
        "ipc",
        "listen",
        "lseek",
        "lstat",
        "lstat64",
        "madvise",
        "mkdir",
        "mkdirat",
        "mmap",
        "mmap2",
        "mprotect",
        "mremap",
        "munmap",
        "nanosleep",
        "newfstatat",
        "open",
        "openat",
        "pause",
        "perf_event_open",
        "pipe",
        "pipe2",
        "poll",
        "ppoll",
        "prctl",
        "pread64",
        "prlimit64",
        "pselect6",
        "read",
        "readlink",
        "readlinkat",
        "recvfrom",
        "recvmmsg",
        "recvmsg",
        "rename",
        "restart_syscall",
        "rmdir",
        "rt_sigaction",
        "rt_sigpending",
        "rt_sigprocmask",
        "rt_sigqueueinfo",
        "rt_sigreturn",
        "rt_sigsuspend",
        "rt_sigtimedwait",
        "rt_tgsigqueueinfo",
        "sched_getaffinity",
        "sched_yield",
        "seccomp",
        "select",
        "semtimedop",
        "send",
        "sendmmsg",
        "sendmsg",
        "sendto",
        "set_robust_list",
        "set_tid_address",
        "setgid",
        "setgid32",
        "setgroups",
        "setgroups32",
        "setns",
        "setrlimit",
        "setsid",
        "setsidaccept4",
        "setsockopt",
        "setuid",
        "setuid32",
        "sigaltstack",
        "socket",
        "socketcall",
        "socketpair",
        "stat",
        "stat64",
        "statfs",
        "sysinfo",
        "umask",
        "uname",
        "unlink",
        "unlinkat",
        "wait4",
        "waitid",
        "waitpid",
        "write"
      ],
      "action": "SCMP_ACT_ALLOW",
      "args": null
    },
    {
      "names": [
        "setns"
      ],
      "action": "SCMP_ACT_ALLOW",
      "args": [
        {
          "index": 1,
          "value": 1073741824,
          "valueTwo": 0,
          "op": "SCMP_CMP_EQ"
        }
      ],
      "comment": "",
      "includes": {},
      "excludes": {}
    }
  ]
}`,
		},
	})
}
