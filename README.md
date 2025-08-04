# 模块构建工具

[NGA SDK](https://github.com/TianwanTW/NGA-SDK) (构建时自动集成 Shell Utils，自动加密`customize.sh`和`nga-utils.sh`)

[模块示例](https://github.com/OOM-WG/ROOT-Module-EG) (使用本工具需要从此使用模板)

## 依赖

- [Android NDK](https://developer.android.google.cn/ndk)
- [Go](https://golang.google.cn)
- [Garble (可选)](https://github.com/burrowers/garble)

## 使用方法 1 (GitHub Action)

1. 前往示例仓库通过模板生成新项目
2. 修改模块项目代码
3. 通过 GitHub Action 手动或自行修改 Workflow 来实现构建模块

## 使用方法 2 (本地构建)

1. 前往示例仓库通过模板生成新项目
2. 将本项目与模块项目克隆至本地
3. 修改模块项目代码
4. 通过 Go 运行本项目的构建工具来实现构建模块

## 项目结构

> [!TIP]
> 模块项目应该遵循以下目录结构

```plaintext
|
├── src                     <--- 模块的目录
│   │
│   │      *** 模块配置文件 ***
│   │
│   ├── module.prop         <--- 此文件用于保存模块相关的一些配置，如模块 ID、版本等
│   │
│   │      *** 可选文件 ***
│   │
│   ├── customize.sh        <--- 此脚本用于控制安装模块时的行为
│   ├── action.sh           <--- 此脚本可在root管理器内通过按钮提供给用户执行
│   │                              (SakitinSU默认支持/Magisk 27008+/KernelSU 1.0.2+/APatch 11039+，低版本不支持该按钮，开发者应避免不支持而是
│   │                               建议使用者脱离啃老或者建议使用者从KernelSU转向KernelSU Next等分支以获取更好的体验(KSU官方版非GKI百分百用不了啦！))
│   ├── post-fs-data.sh     <--- 此脚本将会在 post-fs-data 模式下运行
│   ├── post-mount.sh       <--- 此脚本将会在 post-mount 模式下运行 (仅受KernelSU/APatch支持)
│   ├── service.sh          <--- 此脚本将会在 late_start 服务模式下运行
│   ├── boot-completed.sh   <--- 此脚本将会在 Android 系统启动完毕后以服务模式运行 (仅受KernelSU/APatch支持，Magisk可在“service.sh”内主动执行)
|   ├── uninstall.sh        <--- 此脚本将会在模块被卸载时运行
│   ├── system.prop         <--- 此文件中指定的属性将会在系统启动时通过 resetprop 更改
│   ├── sepolicy.rule       <--- 此文件中的 SELinux 策略将会在系统启动时加载
│   │
│   │      *** 构建时文件 (在构建时会自动添加的文件，不应在您的存储库内出现这些文件) ***
│   │
│   ├── nga-utils.sh        <--- 此文件将会在构建时自动添加，内含各种常用Shell函数
│   ├── hashList.dat        <--- 此文件将会在构建时自动添加，内含模块文件的哈希值，用于安装时校验
│   ├── META-INF            <--- 此目录将会在构建时自动添加，内含模块Recovery刷入脚本 (Magisk模块必须包含这些文件，如果仅支持KernelSU/APatch则无需包含)
│   ├── zygisk              <--- 如果在“c++_native”目录有共享库项目，则会自动添加进此目录，用于Zygisk相关功能 (此目录也可自己添加)
│   ├── bin                 <--- 如果在“c++_native”或“go_native”目录有可执行文件项目，则会自动添加进此目录 (此目录也可自己添加)
│   │
│   └── ...                 <--- 其他自定义文件
│
├── c++_native              <--- C++源码的目录 (仅支持以Android.mk配置的项目)
│   ├── <项目名称>          <--- 项目名称 (必须和“LOCAL_MODULE”值相同，即生成后的二进制名称)
│   │   ├── Android.mk      <--- 构建配置文件 (如果以“BUILD_EXECUTABLE”形式构建，
│   │   │                          将自动把编译后二进制文件移动至模块目录的“bin/<项目名称>/<架构>.bin”，
│   │   │                          如果以“BUILD_SHARED_LIBRARY”形式构建，将自动把编译后二进制文件移动至模块目录的“zygisk”文件夹里面)
│   │   ├── Application.mk  <--- 构建配置文件
│   │   └── ...             <--- 其他源码文件
│   │
│   └── ...                 <--- 其他C++项目
│
├── go_native               <--- Go源码的目录
│   ├── <项目名称>          <--- 项目名称 (必须和“module”值的最后名称相同，即生成后的二进制名称)
│   │   ├── arch            <--- 架构配置目录
│   │   │   ├── go.mod      <--- 架构项目信息文件
│   │   │   ├── go.sum      <--- 有依赖时生成的校验文件
│   │   │   └── ...         <--- 其他源码文件 (用于输出需要构建的架构，按行输出，
│   │   │                          架构以Go支持的为准(例如使用“amd64”而非“x86_64”))
│   │   ├── go.mod          <--- 项目信息文件
│   │   ├── go.sum          <--- 有依赖时生成的校验文件
│   │   └── ...             <--- 其他源码文件
│   └── ...                 <--- 其他Go项目
│
├── README.md               <--- 仓库说明文件
├── changelog.md            <--- 模块更新日志 (可自定义名称/路径)
├── update.json             <--- 用于更新模块的 JSON 文件 (可自定义名称/路径)
│
└── ...                     <--- 其他用于配置Git仓库的文件或自定义文件
```

模块示例仓库已经配置好这些内容，可直接使用模板并按需修改以使用

> [!WARNING]
> 不需要的文件请自行删除！请避免对用户设备造成不必要的负担！
