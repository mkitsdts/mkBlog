[Unit]
Description=__APP_NAME__ blog service
After=network.target

[Service]
Type=simple
User=__USER__
Group=__GROUP__
WorkingDirectory=__ROOT_DIR__
ExecStart=__BIN_PATH__
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
