[Unit]
Description=__APP_NAME__ blog service
After=network.target
StartLimitIntervalSec=60
StartLimitBurst=5

[Service]
Type=simple
User=__USER__
Group=__GROUP__
WorkingDirectory=__ROOT_DIR__
ExecStart=__BIN_PATH__
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
