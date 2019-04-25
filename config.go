package main

var Config = []Directive{
    Deb{
        Name: "nginx",
        Install: true,
        Upgrade: true,
    },
    Deb{
        Name: "php5-fpm",
        Install: true,
        Upgrade: true,
    },
    File{
       Path: "/etc/nginx/sites-available/default",
       Owner: 1000,
       Group: 1000,
       Mode: 0664,
       Directory: false,
       Create: true,
       Content: "server {\n\tlisten 80 default_server;\n\troot /usr/share/nginx/html;\n\tindex index.php;\n\tlocation ~ \\.php$ {\n\t\ttry_files $uri =404;\n\t\tfastcgi_split_path_info ^(.+\\.php)(/.+)$;\n\t\tinclude fastcgi_params;\n\t\tfastcgi_pass unix:/var/run/php5-fpm.sock;\n\t\tfastcgi_index index.php;\n\t\tfastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;\n\t}\n}",
    },
    File{
        Path: "/usr/share/nginx/html/index.php",
        Owner: 1000,
        Group: 1000,
        Mode: 0664,
        Directory: false,
        Create: true,
        Content: "<?php\nheader('Content-Type: text/plain');\necho 'Hello, world!';",
    },
    Service{
        Name: "php5-fpm",
        Running: true,
        Restart: true,
    },
    Service{
        Name: "nginx",
        Running: true,
        Restart: true,
    },
}

