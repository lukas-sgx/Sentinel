package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func copyContent(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	args := []string{"-r"}
	for _, e := range entries {
		args = append(args, filepath.Join(src, e.Name()))
	}
	args = append(args, dst)

	cmd := exec.Command("cp", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runMount(cmd ...string) {
	syscall.Unshare(syscall.CLONE_NEWNS)
	c := exec.Command("mount", cmd...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		panic(err)
	}
}

func runUmount(target string) {
    if target == "" {
        return
    }
    if _, err := os.Stat(target); os.IsNotExist(err) {
        return
    }
    if err := syscall.Unmount(target, syscall.MNT_DETACH); err != nil {
        if err == syscall.ENOENT || err == syscall.EINVAL {
            return
        }
        panic(err)
    }
}

func mountProc(root string) error {
    target := filepath.Join(root, "proc")
    if err := os.MkdirAll(target, 0755); err != nil {
        return err
    }
    return syscall.Mount("proc", target, "proc", syscall.MS_NOSUID|syscall.MS_NOEXEC|syscall.MS_NODEV, "")
}

func init_env() string {
	sandbox, err := os.MkdirTemp("/tmp", "sandbox-root-")
	if err != nil {
		panic(err)
	}
	fmt.Println("[SANDBOX] ->", sandbox)
	fmt.Println("[CP] -> rootfs")
	copyContent("rootfs", sandbox)
	fmt.Println("[USER] -> Sentinel")
    createUserInSandbox(sandbox, "sentinel", 1000, 1000)

	// fmt.Println("[MNT] -> /sys")
	// runMount("-t", "sysfs", "none", sandbox+"/sys")
	fmt.Println("[MNT] -> /dev")
    cmd := exec.Command("unshare", "-Urnm", "--fork", "--pid", "--mount-proc").Run()
    if cmd != nil {
        panic(cmd)
    }

	return sandbox
}

func cleanup(dir string) {
	fmt.Println("[UMNT] -> "+dir+"/proc")
	runUmount(dir+"/proc")
	// fmt.Println("[UMNT] -> "+dir+"/sys")
	// runUmount(dir+"/sys")

	runUmount(dir+"/dev/pts")
	fmt.Println("[UMNT] -> "+dir+"/dev/pts")
    runUmount(dir+"/dev/shm")
	fmt.Println("[UMNT] -> "+dir+"/dev/shm")
	runUmount(dir+"/dev")
	fmt.Println("[UMNT] -> "+dir+"/dev")
	fmt.Println("[RM] -> " + dir)

	defer os.RemoveAll(dir)

}

func pivotRoot(newroot string) error {
    if err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
        return fmt.Errorf("failed make / private: %w", err)
    }

    if err := syscall.Mount(newroot, newroot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
        return fmt.Errorf("failed bind-mount newroot: %w", err)
    }

    putold := filepath.Join(newroot, ".pivot_root")
    if err := os.MkdirAll(putold, 0700); err != nil {
        return fmt.Errorf("failed create putold: %w", err)
    }

    if err := syscall.PivotRoot(newroot, putold); err != nil {
        return fmt.Errorf("pivot_root failed: %w", err)
    }

    if err := syscall.Chdir("/"); err != nil {
        return fmt.Errorf("chdir after pivot_root failed: %w", err)
    }

    old := "/.pivot_root"
    if err := syscall.Unmount(old, syscall.MNT_DETACH); err != nil {
        return fmt.Errorf("unmount old root failed: %w", err)
    }
    if err := os.RemoveAll(old); err != nil {
        return fmt.Errorf("remove old root failed: %w", err)
    }

    return nil
}

func createUserInSandbox(root, username string, uid, gid int) {
    etcDir := filepath.Join(root, "etc")
    if err := os.MkdirAll(etcDir, 0755); err != nil {
        panic(err)
    }

    passwdPath := filepath.Join(etcDir, "passwd")
    passwdLine := fmt.Sprintf("%s:x:%d:%d:Sandbox User:/home/%s:/bin/sh\n", username, uid, gid, username)
    f, err := os.OpenFile(passwdPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        panic(err)
    }
    if _, err := f.WriteString(passwdLine); err != nil {
        f.Close()
        panic(err)
    }
    f.Close()

    groupPath := filepath.Join(etcDir, "group")
    groupLine := fmt.Sprintf("%s:x:%d:\n", username, gid)
    gf, err := os.OpenFile(groupPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        panic(err)
    }
    if _, err := gf.WriteString(groupLine); err != nil {
        gf.Close()
        panic(err)
    }
    gf.Close()

    home := filepath.Join(root, "home", username)
    if err := os.MkdirAll(home, 0755); err != nil {
        panic(err)
    }

    if err := os.Chown(home, uid, gid); err != nil {
        if !os.IsPermission(err) {
            panic(err)
        }
    }
}

func makeNode(path string, mode uint32, major, minor, perm int) {
    dev := int((major << 8) | minor)
    if err := syscall.Mknod(path, mode|uint32(perm), dev); err != nil {
        panic(err)
    }
}

func chroot(dir string) {
    if _, err := os.Stat(filepath.Join(dir, "bin/sh")); err != nil {
        panic(fmt.Errorf("missing /bin/sh inside sandbox: %w", err))
    }

    if err := syscall.Unshare(syscall.CLONE_NEWNS); err != nil {
        panic(err)
    }
    
    os.MkdirAll(dir+"/dev", 0755)
    if err := syscall.Mount("tmpfs", dir+"/dev", "tmpfs", 0, "mode=755"); err != nil {
        panic(err)
    }

    os.MkdirAll(dir+"/dev/pts", 0755)
    if err := syscall.Mount("devpts", dir+"/dev/pts", "devpts", 0, "newinstance,ptmxmode=666,mode=620"); err != nil {
        panic(err)
    }

    os.MkdirAll(dir+"/dev/shm", 0755)
    if err := syscall.Mount("tmpfs", dir+"/dev/shm", "tmpfs", 0, "mode=1777"); err != nil {
        panic(err)
    }

    makeNode(dir+"/dev/null",   syscall.S_IFCHR, 1, 3, 0666)
    makeNode(dir+"/dev/zero",   syscall.S_IFCHR, 1, 5, 0666)
    makeNode(dir+"/dev/full",   syscall.S_IFCHR, 1, 7, 0666)
    makeNode(dir+"/dev/random", syscall.S_IFCHR, 1, 8, 0666)
    makeNode(dir+"/dev/urandom",syscall.S_IFCHR, 1, 9, 0666)
    makeNode(dir+"/dev/tty",    syscall.S_IFCHR, 5, 0, 0666)
    makeNode(dir+"/dev/console",syscall.S_IFCHR, 5, 1, 0600)
    makeNode(dir+"/dev/ptmx",    syscall.S_IFCHR, 5, 2, 0666)



    if err := pivotRoot(dir); err != nil {
        panic(err)
    }

	fmt.Println("[MNT] -> /proc")

	cmd := exec.Command("/bin/sh", "-c",
        "mkdir -p /proc && /bin/mount -t proc -o nosuid,noexec,nodev proc /proc || true; "+
            "hostname sandbox; export PATH=/bin:$PATH;  fastfetch; /bin/bash -i",
    )
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC,
	}

	cmd.Start()
	err := cmd.Wait()

	if err != nil {
		cleanup(dir)
        // panic(err)
    } else {
	    cleanup(dir)
    }
}

func main() {
	if os.Getuid() != 0 {
		println("[FATAL] -> script must be run as root")
		os.Exit(84)
	}
	sandbox := init_env()
	chroot(sandbox)
}
