# Magisk Modules Auto Build

[实用Shell代码](https://github.com/GunRain/SKT-Utils)(构建时自动集成)

[模块示例](https://github.com/GunRain/Magisk-Module-EG)


# 使用方法

fork模块示例，添加自己的代码  
fork本仓库然后运行action


在下方按钮创建token，添加进fork后的本仓库的密钥，密钥名称为`USER_GITHUB_TOKEN`

[![](./img/202412012125310.svg)](https://github.com/settings/tokens)

然后运行action按照需求填写参数  

![alt text](./img/image_2.png)
![alt text](./img/image_1.png)

# 注意事项

> [!TIP]
>模块示例示例仓库应该遵循以下目录结构：nn
```
├── .github
|
├── root                    <--- magisk模块的目录                 
│   │
│   │      *** 模块配置文件 ***
│   │
│   ├── module.prop         <--- 此文件保存模块相关的一些配置，如模块 ID、版本等
│   ├── customize.sh        <--- 此文件控制刷入模块时的安装选项
│   │
│   │      *** 可选文件 ***
│   │
│   ├── post-fs-data.sh     <--- 这个脚本将会在 post-fs-data 模式下运行
│   ├── post-mount.sh       <--- 这个脚本将会在 post-mount 模式下运行
│   ├── service.sh          <--- 这个脚本将会在 late_start 服务模式下运行
│   ├── boot-completed.sh   <--- 这个脚本将会在 Android 系统启动完毕后以服务模式运行
|   ├── uninstall.sh        <--- 这个脚本将会在模块被卸载时运行
│   ├── system.prop         <--- 这个文件中指定的属性将会在系统启动时通过 resetprop 更改
│   ├── sepolicy.rule       <--- 这个文件中的 SELinux 策略将会在系统启动时加载
│   │
├── c++_native
│   ├── .
│   └── .
│ 
├── go_native/go_eg
│   ├── .
│   └── .
│  
├── README.md               <--- 仓库说明文件
├── changelog.md            <--- 模块更新日志
├── update.json             <--- 用于更新模块的 JSON 文件
├── .
├── .

```