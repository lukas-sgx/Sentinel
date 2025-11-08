package main

import (
	"fmt"
	"os"
    "os/exec"
)

func runMount(cmd ...string) {
    c := exec.Command("sudo", append([]string{"mount"}, cmd...)...)
    c.Stdout = os.Stdout
    c.Stderr = os.Stderr
    if err := c.Run(); err != nil {
        panic(err)
    }
}

func runUmount(cmd string) {
    c := exec.Command("sudo", "umount", cmd)
    c.Stdout = os.Stdout
    c.Stderr = os.Stderr
    if err := c.Run(); err != nil {
        panic(err)
    }
}

func init_env() string {
    sandbox, err := os.MkdirTemp("/tmp", "sandbox-root-") 
    if err != nil {
        panic(err)
    }
    fmt.Println("[SANDBOX] ->", sandbox)
    os.MkdirAll(sandbox+"/bin", 0755)
    fmt.Println("[MK] -> /bin")
    os.MkdirAll(sandbox+"/lib", 0755)
    fmt.Println("[MK] -> /lib")
    os.MkdirAll(sandbox+"/lib64", 0755)
    fmt.Println("[MK] -> /lib64")
    os.MkdirAll(sandbox+"/usr", 0755)
    fmt.Println("[MK] -> /usr")
    os.MkdirAll(sandbox+"/proc", 0555)
    fmt.Println("[MK] -> /proc")

    fmt.Println("[MNT] -> /bin")
    runMount("--bind", "/bin", sandbox+"/bin")
    fmt.Println("[MNT] -> /lib")
    runMount("--bind", "/lib", sandbox+"/lib")
    fmt.Println("[MNT] -> /lib64")
    runMount("--bind", "/lib64", sandbox+"/lib64")
    fmt.Println("[MNT] -> /usr")
    runMount("--bind", "/usr", sandbox+"/usr")
    fmt.Println("[MNT] -> /proc")
    runMount("-t", "proc", "proc", sandbox+"/proc")


    return sandbox
}

func cleanup(dir string) {
    fmt.Println("[UMNT] -> /bin")
    runUmount(dir+"/bin")
    defer os.RemoveAll(dir)
}

func main() {
    if os.Getuid() != 0 {
        println("[FATAL] -> script must be run as root")
        os.Exit(84)
    }
    sandbox := init_env()
    cleanup(sandbox)
}