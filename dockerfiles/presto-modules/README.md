
# 说明
目前经过拆分后，原presto-search模块分为rfifind,dedisp,search-fold三个模块。各模块功能如下：

- rfifind: fits -> mask 找出射电干扰。输入消息为待处理fits文件所在目录路径，可以一次处理连续的多个fits文件。
- dedisp: fits + mask -> dat 根据指定的LINENUM环境变量执行一组消色散。目前的方案将所有DM分为9组执行，共6041个DM值。输入消息为待处理fits文件所在目录路径，可以一次处理连续的多个fits文件。需要指定LINENUM环境变量(1~9)。处理结果为名为${LINENUM}.tar.zst的压缩包。
- search-fold: 完成以下步骤
    - dat -> fft
    - fft -> cand
    - cand + dat -> 图片（产生结果）

    输入消息为待处理.dat文件压缩包的路径。处理结果为若干干.ps/.pfd/.png文件，和一个包含筛选后结果的candidates.txt文件。

- presto-search-raw: 最早版本的脚本。输入消息为单个fits.zst文件路径，完成整个presto模块的处理流程，并使用原始fits数据执行prepfold。
- presto-search-dat: 输入消息为单个fits.zst文件路径，完成整个presto模块的处理流程，并使用消色散后得到的dat文件执行prepfold。
