# Magisk Modules Auto Build

[实用Shell代码](https://github.com/GunRain/SKT-Utils)(构建时自动集成)

[模块示例](https://github.com/GunRain/Magisk-Module-EG)


# 使用方法

fork模块示例，添加自己的代码

fork本仓库然后运行action


在下方按钮创建token，添加进fork后的本仓库的密钥，密钥名称为`USER_GITHUB_TOKEN`

[![](./img/gh.svg)](https://github.com/settings/tokens)

然后运行action按照需求填写参数  

![](https://img.nightrainmilkyway.cn/img/202501251135576.png)  
![](https://img.nightrainmilkyway.cn/img/202501251135538.png)

# 注意事项

> [!TIP]
>模块示例示例仓库应该遵循以下目录结构:
```
|
├── root                    <--- Magisk模块的目录                 
│   │
│   │      *** 模块配置文件 ***
│   │
│   ├── module.prop         <--- 此文件保存模块相关的一些配置，如模块 ID、版本等
│   ├── customize.sh        <--- 此文件控制刷入模块时的安装选项
│   │
│   │      *** 可选文件 ***
│   │
│   ├── action.sh           <--- 这个脚本可在root管理器内通过按钮提供给用户执行
│   ├── post-fs-data.sh     <--- 这个脚本将会在 post-fs-data 模式下运行
│   ├── post-mount.sh       <--- 这个脚本将会在 post-mount 模式下运行 (仅受KernelSU/APatch支持)
│   ├── service.sh          <--- 这个脚本将会在 late_start 服务模式下运行
│   ├── boot-completed.sh   <--- 这个脚本将会在 Android 系统启动完毕后以服务模式运行 (仅受KernelSU/APatch支持，Magisk可在“service.sh”内主动执行)
|   ├── uninstall.sh        <--- 这个脚本将会在模块被卸载时运行
│   ├── system.prop         <--- 这个文件中指定的属性将会在系统启动时通过 resetprop 更改
│   ├── sepolicy.rule       <--- 这个文件中的 SELinux 策略将会在系统启动时加载
│   │
│   │      *** 构建时文件 (在构建时会自动添加的文件，不应在您的存储库内出现这些文件) ***
│   │
│   ├── skt-utils.sh        <--- 此文件将会在构建时自动添加，内含各种常用Shell函数
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
│   │   ├── arch.txt        <--- 架构配置文件 (以空格分割架构)
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