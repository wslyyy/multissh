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
		multissh -t cmd  -h host -P port(default 22) -u user(default root) -p passswrod [-f] command 
		Files-transfer:   
		<push file>   
		multissh -t push  -h host -P port(default 22) -u user(default root) -p passswrod [-f] localfile  remotepath 
		<pull file> 
		multissh -t pull -h host -P port(default 22) -u user(default root) -p passswrod [-f] remotefile localpath 
	2.Batch Mode
		Ssh-comand:
		multissh -t cmd -i ip_filename -P port(default 22) -u user(default root) -p passswrod [-f] command 
		Files-transfer:   
		multissh -t push -i ip_filename -P port(default 22) -u user(default root) -p passswrod [-f] localfile  remotepath 
		multissh -t pull -i ip_filename -P port(default 22) -u user(default root) -p passswrod [-f] remotefile localpath
EMAIL
    	wusiyuan@lollitech.com 
`
