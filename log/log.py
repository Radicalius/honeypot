import json
import time
import datetime
import boto3
import traceback

class Logger:

    def __init__(self, application, context={}):
        self.context = context
        self.client = boto3.client('logs')
        self.log_group = application
        self.log_stream = str(int(time.time() * 1000))

        try:
            self.client.create_log_group(logGroupName=self.log_group)
        except:
            pass

        try:
            self.client.create_log_stream(logGroupName=self.log_group, logStreamName=self.log_stream)
        except:
            pass

    def bind(self, **kwargs):
        self.context = {**self.context, **kwargs}

    def _log(self, level, **kwargs):
        ts = int(time.time() * 1000)
        kwargs['timestamp'] = datetime.datetime.now().isoformat()
        kwargs['level'] = level
        event = {
            'timestamp': ts,
            'message': json.dumps({**self.context, **kwargs})
        }
        self.client.put_log_events(
            logGroupName=self.log_group,
            logStreamName=self.log_stream,
            logEvents = [event]
        )

    def verbose(self, **kwargs):
        self._log("VERBOSE", **kwargs)

    def debug(self, **kwargs):
        self._log("DEBUG", **kwargs)
    
    def info(self, **kwargs):
        self._log("INFO", **kwargs)

    def warning(self, **kwargs):
        self._log("WARNING", **kwargs)

    def error(self, **kwargs):
        self._log("ERROR", **kwargs)

    def exception(self, **kwargs):
        kwargs['traceback'] = traceback.format_exc()
        self._log("ERROR", **kwargs)
