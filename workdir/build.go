//====================================================================================================
// Copyright (C) 2016-present Anne Sakitin (Tianwan Ayana).                                          =
//                                                                                                   =
// Licensed under the F2DLPR License.                                                                =
//                                                                                                   =
// YOU MAY NOT USE THIS FILE EXCEPT IN COMPLIANCE WITH THE LICENSE.                                  =
// Provided "AS IS", WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,                                   =
// unless required by applicable law or agreed to in writing.                                        =
//                                                                                                   =
// For details about the F2DLPR License terms and conditions, visit: http://license.fileto.download. =
//====================================================================================================

package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"app.niggergo.work/sdk/nga"
	"github.com/ulikunitz/xz"
)

type strs []string

func (_strs *strs) Set(v string) error { *_strs = append(*_strs, v); return nil }
func (_strs *strs) String() string     { return strings.Join(*_strs, ",") }

var (
	cmds  = []string{"go", "ndk-build"}
	err   error
	time0 = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
)

func ndk_build(cmds ...string) bool {
	_, err = exec.Command("ndk-build", cmds...).CombinedOutput()
	return err == nil
}

func go_env(envs map[string]string) bool {
	for key, val := range envs {
		_, err = exec.Command("go", "env", "-w", key+"="+val).CombinedOutput()
		if err != nil {
			return false
		}
	}
	return true
}

func go_build(garble bool) bool {
	var cmd []string
	if garble {
		cmd = []string{"garble", "-literals", "-seed=random", "-tiny"}
	} else {
		cmd = []string{"go"}
	}
	cmd = append(cmd, "build", "-ldflags=-w -s")
	_, err = exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	return err == nil
}

func cpp_arch2arch(arch string) string {
	switch arch {
	case "arm64-v8a":
		return "arm64"
	case "armeabi-v7a":
		return "arm"
	case "x86_64":
		return "x64"
	default:
		return arch
	}
}

func go_arch2arch(arch string) string {
	switch arch {
	case "amd64":
		return "x64"
	case "386":
		return "x86"
	default:
		return arch
	}
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("[!] Error: \tcannot get workdir")
		return
	}

	in := flag.String("in", filepath.Join(wd, "input"), "input dir")
	out := flag.String("out", filepath.Join(wd, "output"), "output dir")
	var names strs
	flag.Var(&names, "names", "module names")
	out_name := flag.String("outname", "instpkg", "output name")
	garble := flag.Bool("garble", true, "use garble")
	upx := flag.Bool("upx", false, "use upx")
	xz_zip := flag.Bool("xz", false, "use xz")

	flag.Parse()

	if *garble {
		cmds = append(cmds, "garble")
	}
	if *upx {
		cmds = append(cmds, "upx")
	}

	for _, cmd := range cmds {
		if _, err = exec.LookPath(cmd); err != nil {
			fmt.Printf("[!] Error: \tcommand \"%s\" not found\n", cmd)
			return
		}
	}
	var shell string
	if _, err = exec.LookPath("bash"); err != nil {
		if _, err = exec.LookPath("sh"); err != nil {
			fmt.Println("[!] Error: \tcommand \"bash\" or \"sh\" not found")
			return
		} else {
			shell = "sh"
		}
	} else {
		shell = "bash"
	}

	if !nga.PathExist(*in) || !nga.IsDir(*in) {
		fmt.Printf("[!] Error: \tpath \"%s\" cannot be used\n", *in)
		return
	}
	if !nga.PathExist(*out) {
		err = os.MkdirAll(*out, os.ModePerm)
		if err != nil {
			fmt.Printf("[!] Error: \tcannot create dir \"%s\"\n", *out)
			return
		}
	} else if !nga.IsDir(*out) {
		fmt.Printf("[!] Error: \tpath \"%s\" cannot be used\n", *out)
		return
	}

	var mods []string
	if len(names) == 0 {
		entries, err := os.ReadDir(*in)
		if err != nil {
			fmt.Printf("[!] Error: \tcannot read dir \"%s\"\n", *in)
			return
		}
		for _, entry := range entries {
			if nga.PathExist(filepath.Join(*in, entry.Name(), "root", "module.prop")) {
				mods = append(mods, entry.Name())
				fmt.Printf("[+] Added: \tModule \"%s\" to Build List\n", entry.Name())
			}
		}
	} else {
		for _, name := range names {
			if nga.PathExist(filepath.Join(*in, name, "root", "module.prop")) {
				mods = append(mods, name)
				fmt.Printf("[+] Added: \tModule \"%s\" to Build List\n", name)
			}
		}
	}

	tmp_dir := filepath.Join(*out, "._mod_bld_tmp")
	if nga.PathExist(tmp_dir) {
		err = os.RemoveAll(tmp_dir)
		if err != nil {
			fmt.Printf("[!] Error: \tcannot delete dir \"%s\"\n", tmp_dir)
			return
		}
	}
	for _, mod := range mods {
		mod_dir := filepath.Join(*in, mod)
		fmt.Printf("[*] Building: \tModule \"%s\"\n", mod)
		err = os.Chdir(wd)
		if err != nil {
			fmt.Printf("[!] Error: \tcannot change workdir to \"%s\"\n", wd)
			return
		}
		err = os.MkdirAll(tmp_dir, os.ModePerm)
		if err != nil {
			fmt.Printf("[!] Error: \tcannot create dir \"%s\"\n", tmp_dir)
			return
		}
		if cpp_dir := filepath.Join(mod_dir, "c++_native"); nga.PathExist(cpp_dir) {
			entries, err := os.ReadDir(cpp_dir)
			if err != nil {
				fmt.Printf("[!] Error: \tcannot read dir \"%s\"\n", cpp_dir)
				return
			}
			for _, entry := range entries {
				bin_name := entry.Name()
				fmt.Printf("[*] Building: \tC++ Binary \"%s\"\n", bin_name)
				jni_dir := filepath.Join(cpp_dir, bin_name, "jni")
				lib_dir := filepath.Join(cpp_dir, bin_name, "libs")
				mk_path := filepath.Join(jni_dir, "Android.mk")
				if !nga.PathExist(mk_path) ||
					!nga.PathExist(filepath.Join(jni_dir, "Application.mk")) {
					continue
				}
				err = os.Chdir(jni_dir)
				if err != nil {
					fmt.Printf("[!] Error: \tcannot change workdir to \"%s\"\n", jni_dir)
					return
				}
				var zygisk bool
				mk_dat, err := os.ReadFile(mk_path)
				if err != nil {
					fmt.Printf("[!] Error: \tcannot read file \"%s\"\n", mk_path)
					return
				}
				if strings.Contains(string(mk_dat), "\ninclude $(BUILD_SHARED_LIBRARY)") {
					zygisk = true
				}
				ndk_build("-j" + strconv.Itoa(runtime.NumCPU()))
				if ndk_build() {
					if zygisk {
						fmt.Printf("[✓] Built: \tC++ Zygisk Library \"%s\"\n", bin_name)
					} else {
						fmt.Printf("[✓] Built: \tC++ Executable \"%s\"\n", bin_name)
					}
				} else {
					if zygisk {
						fmt.Printf("[!] Error: \tcannot build c++ zygisk library \"%s\"\n", bin_name)
					} else {
						fmt.Printf("[!] Error: \tcannot build c++ executable \"%s\"\n", bin_name)
					}
					return
				}
				if zygisk_dir := filepath.Join(tmp_dir, "zygisk"); zygisk {
					entries, err := os.ReadDir(lib_dir)
					if err != nil {
						fmt.Printf("[!] Error: \tcannot read dir \"%s\"\n", lib_dir)
						return
					}
					err = os.MkdirAll(zygisk_dir, os.ModePerm)
					if err != nil {
						fmt.Printf("[!] Error: \tcannot create dir \"%s\"\n", zygisk_dir)
						return
					}
					for _, entry := range entries {
						target_arch := entry.Name()
						target_bin_name := "lib" + bin_name + ".so"
						if target_bin := filepath.Join(lib_dir, target_arch, target_bin_name); nga.PathExist(target_bin) {
							if err = nga.MoveFile(target_bin, filepath.Join(zygisk_dir, target_arch+".so")); err != nil {
								fmt.Printf("[!] Error: \tcannot move c++ zygisk library \"%s\" (arch: %s)\n", bin_name, target_arch)
								return
							} else {
								fmt.Printf("[→] Moved: \tC++ Zygisk Library \"%s\" (Arch: %s)\n", bin_name, target_arch)
							}
						}
					}
				} else {
					entries, err := os.ReadDir(lib_dir)
					if err != nil {
						fmt.Printf("[!] Error: \tcannot read dir \"%s\"\n", lib_dir)
						return
					}
					for _, entry := range entries {
						target_arch := entry.Name()
						target_exe_dir := filepath.Join(tmp_dir, "bin", bin_name)
						if target_bin := filepath.Join(lib_dir, target_arch, bin_name); nga.PathExist(target_bin) {
							err = os.MkdirAll(target_exe_dir, os.ModePerm)
							if err != nil {
								fmt.Printf("[!] Error: \tcannot create dir \"%s\"\n", target_exe_dir)
								return
							}
							if err = nga.MoveFile(target_bin, filepath.Join(target_exe_dir, cpp_arch2arch(target_arch)+".elf")); err != nil {
								fmt.Printf("[!] Error: \tcannot move c++ executable \"%s\" (arch: %s)\n", bin_name, target_arch)
								return
							} else {
								fmt.Printf("[→] Moved: \tC++ Executable \"%s\" (Arch: %s)\n", bin_name, target_arch)
							}
						}
					}
				}
			}
		}
		if go_dir := filepath.Join(mod_dir, "go_native"); nga.PathExist(go_dir) {
			entries, err := os.ReadDir(go_dir)
			if err != nil {
				fmt.Printf("[!] Error: \tcannot read dir \"%s\"\n", go_dir)
				return
			}
			for _, entry := range entries {
				bin_name := entry.Name()
				fmt.Printf("[*] Building: \tGo Executable \"%s\"\n", bin_name)
				bin_dir := filepath.Join(go_dir, bin_name)
				arch_dir := filepath.Join(bin_dir, "arch")
				if !nga.PathExist(filepath.Join(bin_dir, "go.mod")) ||
					!nga.PathExist(filepath.Join(arch_dir, "go.mod")) {
					continue
				}
				err = os.Chdir(arch_dir)
				if err != nil {
					fmt.Printf("[!] Error: \tcannot change workdir to \"%s\"\n", bin_dir)
					return
				}
				if !go_env(map[string]string{
					"GOOS":   runtime.GOOS,
					"GOARCH": runtime.GOARCH,
				}) {
					fmt.Println("[!] Error: \tcannot set go env")
					return
				}
				archs, err := exec.Command("go", "run", ".").CombinedOutput()
				if err != nil {
					fmt.Println("[!] Error: \tcannot get arch")
					return
				}
				err = os.Chdir(bin_dir)
				if err != nil {
					fmt.Printf("[!] Error: \tcannot change workdir to \"%s\"\n", bin_dir)
					return
				}
				if !go_env(map[string]string{
					"GOOS":        "linux",
					"CGO_ENABLED": "0",
				}) {
					fmt.Println("[!] Error: \tcannot set go env")
					return
				}
				scanner := bufio.NewScanner(strings.NewReader(string(archs)))
				for scanner.Scan() {
					target_arch := strings.TrimSpace(scanner.Text())
					if target_arch == "" {
						continue
					}
					if !go_env(map[string]string{"GOARCH": target_arch}) {
						fmt.Println("[!] Error: \tcannot set go env")
						return
					}
					if go_build(*garble) {
						fmt.Printf("[✓] Built: \tGo Executable \"%s\" (Arch: %s)\n", bin_name, target_arch)
					} else {
						fmt.Printf("[!] Error: \tcannot build go executable \"%s\" (arch: %s)\n", bin_name, target_arch)
						return
					}
					target_exe_dir := filepath.Join(tmp_dir, "bin", bin_name)
					if target_bin := filepath.Join(bin_dir, bin_name); nga.PathExist(target_bin) {
						err = os.MkdirAll(target_exe_dir, os.ModePerm)
						if err != nil {
							fmt.Printf("[!] Error: \tcannot create dir \"%s\"\n", target_exe_dir)
							return
						}
						if err = nga.MoveFile(target_bin, filepath.Join(target_exe_dir, go_arch2arch(target_arch)+".elf")); err != nil {
							fmt.Printf("[!] Error: \tcannot move go executable \"%s\" (arch: %s)\n", bin_name, target_arch)
							return
						} else {
							fmt.Printf("[→] Moved: \tGo Executable \"%s\" (Arch: %s)\n", bin_name, target_arch)
						}
					}
				}
				if err = scanner.Err(); err != nil {
					fmt.Println("[!] Error: \tcannot scan archs")
					return
				}
				if !go_env(map[string]string{
					"GOOS":   runtime.GOOS,
					"GOARCH": runtime.GOARCH,
				}) {
					fmt.Println("[!] Error: \tcannot set go env")
					return
				}
			}
		}
		err = os.Chdir(wd)
		if err != nil {
			fmt.Printf("[!] Error: \tcannot change workdir to \"%s\"\n", wd)
			return
		}

		if err = nga.CopyDir(
			filepath.Join(wd, "src", "META-INF"),
			filepath.Join(tmp_dir, "META-INF"),
		); err != nil {
			fmt.Println("[!] Error: \tcannot copy recovery flash script")
			return
		} else {
			fmt.Printf("[=] Copied: \tRecovery Flash Script for Module \"%s\"\n", mod)
		}

		nga_dir := filepath.Join(wd, "src", "nga-sdk", "src", "shell")
		if err = nga.CopyFile(
			filepath.Join(nga_dir, "nga-utils.sh"),
			filepath.Join(tmp_dir, "nga-utils.sh"),
		); err != nil {
			fmt.Println("[!] Error: \tcannot copy nga shell utils")
			return
		} else {
			fmt.Printf("[=] Copied: \tNGA Shell Utils for Module \"%s\"\n", mod)
		}

		root_dir := filepath.Join(mod_dir, "root")
		if err = nga.CopyDir(root_dir, tmp_dir); err != nil {
			fmt.Printf("[!] Error: \tcannot copy module \"%s\" files\n", mod)
			return
		} else {
			fmt.Printf("[=] Copied: \tModule \"%s\" Files\n", mod)
		}

		prop_path := filepath.Join(tmp_dir, "module.prop")
		prop_dat, err := os.ReadFile(filepath.Join(tmp_dir, "module.prop"))
		if err != nil {
			fmt.Printf("[!] Error: \tcannot read file \"%s\"\n", prop_path)
			return
		}
		if strings.Contains(string(prop_dat), "咲汀") ||
			strings.Contains(string(prop_dat), "Sakitin") ||
			strings.Contains(string(prop_dat), "OOM. WG.") {
			if err = nga.CopyFile(
				filepath.Join(filepath.Dir(wd), "LICENSE.txt"),
				filepath.Join(tmp_dir, "LICENSE.txt"),
			); err != nil {
				fmt.Println("[!] Error: \tcannot copy F2DLPR License")
				return
			} else {
				fmt.Printf("[=] Copied: \tF2DLPR License for Module \"%s\"\n", mod)
			}
		}

		if *upx {
			if err = filepath.WalkDir(filepath.Join(tmp_dir, "bin"), func(path string, dir fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !dir.IsDir() {
					if _, err = exec.Command("upx", "--best", path).CombinedOutput(); err != nil {
						fmt.Printf("[!] Error: \tcannot use upx for executable \"%s\"\n", filepath.Base(path))
					} else {
						fmt.Printf("[✓] Compressed: \tExecutable \"%s\" by UPX\n", filepath.Base(path))
					}
				}
				return nil
			}); err != nil {
				fmt.Println("[!] Error: \tcannot use upx")
				return
			}

		}

		enc_path, err := filepath.Rel(wd, filepath.Join(nga_dir, "nga-enc.sh"))
		enc_path = filepath.ToSlash(enc_path)
		if err != nil {
			fmt.Printf("[!] Error: \tcannot get relative path for \"%s\"\n", filepath.Join(nga_dir, "nga-enc.sh"))
			return
		}
		utils_path, err := filepath.Rel(wd, filepath.Join(tmp_dir, "nga-utils.sh"))
		utils_path = filepath.ToSlash(utils_path)
		if err != nil {
			fmt.Printf("[!] Error: \tcannot get relative path for \"%s\"\n", filepath.Join(nga_dir, "nga-utils.sh"))
			return
		}
		cust_path, err := filepath.Rel(wd, filepath.Join(tmp_dir, "customize.sh"))
		cust_path = filepath.ToSlash(cust_path)
		if err != nil {
			fmt.Printf("[!] Error: \tcannot get relative path for \"%s\"\n", filepath.Join(nga_dir, "customize.sh"))
			return
		}
		if nga.PathExist(utils_path) {
			if _, err = exec.Command(shell,
				enc_path,
				utils_path,
			).CombinedOutput(); err != nil {
				fmt.Printf("[!] Error: \tcannot encrypt script \"%s\"\n", "nga-utils.sh")
				return
			} else {
				fmt.Printf("[$] Encrypted: \tScript \"%s\"\n", "nga-utils.sh")
			}
		}
		if nga.PathExist(cust_path) {
			if _, err = exec.Command(shell,
				enc_path,
				cust_path,
			).CombinedOutput(); err != nil {
				fmt.Printf("[!] Error: \tcannot encrypt script \"%s\"\n", "customize.sh")
				return
			} else {
				fmt.Printf("[$] Encrypted: \tScript \"%s\"\n", "customize.sh")
			}
		}

		var buffer bytes.Buffer
		if err = filepath.Walk(tmp_dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || strings.Contains(path, "META-INF") {
				return nil
			}
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			hash1 := sha512.New384()
			if _, err := io.Copy(hash1, file); err != nil {
				return err
			}
			hash2 := sha1.New()
			hash2.Write([]byte(hex.EncodeToString(hash1.Sum(nil))))
			rel, err := filepath.Rel(tmp_dir, path)
			if err != nil {
				return err
			}
			buffer.WriteString(hex.EncodeToString(hash2.Sum(nil)) + " " + filepath.ToSlash(rel) + "\n")
			return nil
		}); err != nil {
			fmt.Println("[!] Error: \tcannot get hashes")
			return
		}
		hashes_dat := buffer.Bytes()
		if len(hashes_dat) > 0 && hashes_dat[len(hashes_dat)-1] == '\n' {
			hashes_dat = hashes_dat[:len(hashes_dat)-1]
		}
		var encoded bytes.Buffer
		gz := gzip.NewWriter(&encoded)
		b64 := base64.NewEncoder(base64.StdEncoding, gz)
		if _, err = b64.Write(hashes_dat); err != nil {
			fmt.Println("[!] Error: \tcannot write base64")
			return
		}
		if err = b64.Close(); err != nil {
			fmt.Println("[!] Error: \tcannot close base64")
			return
		}
		if err = gz.Close(); err != nil {
			fmt.Println("[!] Error: \tcannot close gunzip")
			return
		}
		if err = os.WriteFile(filepath.Join(tmp_dir, "hashList.dat"), encoded.Bytes(), os.ModePerm); err != nil {
			fmt.Println("[!] Error: \tcannot write hashes")
			return
		} else {
			fmt.Printf("[✓] Wrote: \tModule \"%s\" File Hashes\n", mod)
		}

		out_dir := filepath.Join(*out, mod)
		if err = os.MkdirAll(out_dir, os.ModePerm); err != nil {
			fmt.Printf("[!] Error: \tcannot create module \"%s\" output dir\n", mod)
			return
		} else {
			fmt.Printf("[+] Created: \tModule \"%s\" Output Dir\n", mod)
		}
		zip_name := *out_name + ".zip"
		out_path := filepath.Join(out_dir, zip_name)
		zip_file, err := os.Create(out_path)
		if err != nil {
			fmt.Printf("[!] Error: \tcannot create module \"%s\" output zip\n", mod)
			return
		}
		defer zip_file.Close()
		zip_writer := zip.NewWriter(zip_file)
		zip_writer.RegisterCompressor(zip.Deflate, func(w io.Writer) (io.WriteCloser, error) {
			return flate.NewWriter(w, flate.BestCompression)
		})
		defer zip_writer.Close()
		if err = filepath.WalkDir(tmp_dir, func(path string, dir os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if dir.IsDir() {
				return nil
			}
			rel_path, err := filepath.Rel(tmp_dir, path)
			if err != nil {
				return err
			}
			rel_path = filepath.ToSlash(rel_path)
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			info, err := dir.Info()
			if err != nil {
				return err
			}
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = rel_path
			header.Method = zip.Deflate
			header.Modified = time0
			writer, err := zip_writer.CreateHeader(header)
			if err != nil {
				return err
			}
			_, err = io.Copy(writer, file)
			return err
		}); err != nil {
			fmt.Printf("[!] Error: \tcannot create module \"%s\" output zip \"%s\"\n", mod, zip_name)
			return
		} else {
			fmt.Printf("[+] Created: \tModule \"%s\" Output Zip \"%s\"\n", mod, zip_name)
		}
		if *xz_zip {
			zip_name := *out_name + "_xz.zip"
			out_path := filepath.Join(out_dir, zip_name)
			zip_file, err := os.Create(out_path)
			if err != nil {
				fmt.Printf("[!] Error: \tcannot create module \"%s\" output zip (xz)\n", mod)
				return
			}
			defer zip_file.Close()
			zip_writer := zip.NewWriter(zip_file)
			zip_writer.RegisterCompressor(95, func(w io.Writer) (io.WriteCloser, error) {
				return xz.WriterConfig{
					DictCap: 1 << 26,
				}.NewWriter(w)
			})
			defer zip_writer.Close()
			if err = filepath.WalkDir(tmp_dir, func(path string, dir os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if dir.IsDir() {
					return nil
				}
				rel_path, err := filepath.Rel(tmp_dir, path)
				if err != nil {
					return err
				}
				rel_path = filepath.ToSlash(rel_path)
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				info, err := dir.Info()
				if err != nil {
					return err
				}
				header, err := zip.FileInfoHeader(info)
				if err != nil {
					return err
				}
				header.Name = rel_path
				header.Method = 95
				header.Modified = time0
				writer, err := zip_writer.CreateHeader(header)
				if err != nil {
					return err
				}
				_, err = io.Copy(writer, file)
				return err
			}); err != nil {
				fmt.Printf("[!] Error: \tcannot create module \"%s\" output zip \"%s\" (xz)\n", mod, zip_name)
				return
			} else {
				fmt.Printf("[+] Created: \tModule \"%s\" Output Zip \"%s\" (xz)\n", mod, zip_name)
			}
		}
		if err = os.RemoveAll(tmp_dir); err != nil {
			fmt.Printf("[!] Error: \tcannot clean module \"%s\" build cache\n", mod)
			return
		} else {
			fmt.Printf("[-] Cleaned: \tModule \"%s\" Build Cache\n", mod)
		}
	}
}
