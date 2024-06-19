#!/usr/bin/env bash
echo $LOG_PATH
exec ./main -cfg config &>> $LOG_PATH