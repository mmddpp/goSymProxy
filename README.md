# goSymProxy

Microsoft Symbol Server Proxy implemented in pure Go.

You can deploy goSymProxy independently without requiring **IIS** or **SymProxy.dll**.

## config

ip: listen address  
port: listen port  
root: path to cache symbol files  
route: proxy URI
timeout: connect timeout seconds  

## test link

source link:  
<http://msdl.microsoft.com/download/symbols/wntdll.pdb/F999943DF7FB4B8EB6D99F2B047BC3101/wntdll.pdb>

proxy link:  
<http://127.0.0.1/download/symbols/wntdll.pdb/F999943DF7FB4B8EB6D99F2B047BC3101/wntdll.pdb>
