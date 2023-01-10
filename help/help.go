package help

const Help = `			 multissh
NAME
	multissh is a smart ssh tool.It is developed by Go,compiled into a separate binary without any dependencies.
DESCRIPTION
		multissh can do the follow things:
		1.runs cmd on the remote host.
		2.push a local file or path to the remote host.
		3.pull remote host file to local.
USAGE
	1.Single Mode
		remote-comand:
		multissh -t cmd(default cmd)  -h host -P port(default 22) -u user(default root) [-p passswrod] [-K privateSSHKeyPath] [-f] command 
		Files-transfer:   
		<push file>   
		multissh -t push  -h host -P port(default 22) -u user(default root) -p [-p passswrod] [-K privateSSHKeyPath] [-f] localfile  remotepath 
		<pull file> 
		multissh -t pull -h host -P port(default 22) -u user(default root) -p [-p passswrod] [-K privateSSHKeyPath] [-f] remotefile localpath 
	2.Batch Mode
		Ssh-comand:
		multissh -t cmd(default cmd) -i ipFilePath -P port(default 22) -u user(default root) -p passswrod [-f] command 
		Files-transfer:   
		multissh -t push -i ipFilePath -P port(default 22) -u user(default root) [-f] localfile  remotepath 
		multissh -t pull -i ipFilePath -P port(default 22) -u user(default root) [-f] remotefile localpath
IPFILE
	Create ipfile.txt, Every line is like:
		mode[k/p]|host|port|user|[privateSSHKeyPath][password]
	example:
		p|192.168.88.36|22|root|PWD!123456
		k|192.168.88.26|22|root|/root/.ssh/a.key
		k|192.168.88.16|22|root|/root/.ssh/b.key
EMAIL
    	wusiyuan@lollitech.com
`
