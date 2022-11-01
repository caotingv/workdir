import logging
import os
import webob

from io import BytesIO
from flask import Flask, request, send_file
from werkzeug import serving, secure_filename


SMB_FOLDER = '/data/fsd/111/'
LOG_DATEFMT = "%Y-%m-%d %H:%M:%S"
LOG_FORMAT = "%(asctime)s.%(msecs)03d - %(levelname)s - %(message)s"
logging.basicConfig(format=LOG_FORMAT, datefmt=LOG_DATEFMT)
LOG = logging.getLogger(__name__)
LOG.setLevel(logging.INFO)


class AgentServrer(object):

    def __init__(self):
        self.app = Flask(__name__)
        self.app.add_url_rule(rule='/',
                              view_func=self.get_rootpath_file)
        self.app.add_url_rule(rule='/<path:subpath>',
                              methods=['GET', 'POST', 'DELETE'],
                              view_func=self.folder_events)
        self.app.add_url_rule(rule='/upload', methods=['POST'],
                              view_func=self.upload_file)
        self.app.add_url_rule(rule='/download',
                              view_func=self.download_file)
        self.app.add_url_rule(rule='/delete',
                              methods=['DELETE'],
                              view_func=self.delete_file)

    # curl http://127.0.0.1:5000 | json_pp
    def get_rootpath_file(self):
        files = []

        for filename in os.listdir(SMB_FOLDER):
            abs_path = os.path.join(SMB_FOLDER, filename)
            stat = os.stat(abs_path)
            files.append({
                'filename': filename,
                'create_time': stat.st_ctime,
                'update_time': stat.st_mtime,
                'owner_id': stat.st_uid,
                'group_id': stat.st_gid,
                'file_size': stat.st_size,
                'mode': stat.st_mode,
                'is_directory': os.path.isdir(abs_path),
                'is_file': os.path.isfile(abs_path)
            })

        return webob.Response(json={'files': files})

    # curl http://127.0.0.1:5000/aaa | json_pp
    def folder_events(self, subpath):
        if request.method == 'GET':
            files = []

            file_path = os.path.join(SMB_FOLDER + subpath)
            for filename in os.listdir(file_path):
                abs_path = os.path.join(file_path, filename)
                stat = os.stat(abs_path)
                files.append({
                    'filename': filename,
                    'create_time': stat.st_ctime,
                    'update_time': stat.st_mtime,
                    'owner_id': stat.st_uid,
                    'group_id': stat.st_gid,
                    'file_size': stat.st_size,
                    'mode': stat.st_mode,
                    'is_directory': os.path.isdir(abs_path),
                    'is_file': os.path.isfile(abs_path)
                })

            return webob.Response(json={'files': files})

        elif request.method == 'POST':
            LOG.info('The create folder is : %s', subpath)

            abs_path = os.path.join(SMB_FOLDER, subpath)
            os.mkdir(abs_path)

            return webob.Response(json={"accept": True})

        elif request.method == 'DELETE':
            LOG.info('The delete folder is : %s', subpath)

            abs_path = os.path.join(SMB_FOLDER, subpath)
            os.removedirs(abs_path)

            return webob.Response(json={"accept": True})

    # curl -L -X POST  http://127.0.0.1:5000/upload?path=aaa
    # -H 'Content-Type: multipart/form-data'
    # -F file=@'/root/test'
    def upload_file(self):
        file = request.files.get('file')
        path = request.args.get('path')
        LOG.info('Upload file %s to %s ', file, path)

        upload_path = SMB_FOLDER
        if path:
            upload_path = os.path.join(SMB_FOLDER, path)

        filename = secure_filename(file.filename)
        upload_file = os.path.join(upload_path, filename)

        if os.path.exists(upload_file):
            pass
        else:
            file.save(upload_file)

        return webob.Response(json={"accept": True})

    # curl -O -L -X GET  http://127.0.0.1:5000/download?file=
    def download_file(self):
        file = request.args.get('file')
        LOG.info('The download file is : %s', file)

        download_file = os.path.join(SMB_FOLDER, file)

        if os.path.exists(download_file):
            with open(download_file, 'rb') as bites:
                return send_file(
                    BytesIO(bites.read()),
                    download_name=file,
                    as_attachment=True
                )
        else:
            return webob.Response(json={"error": 'No such file on server'}, status=500)

    # curl -L -X DELETE  http://127.0.0.1:5000/delete?file=bbb.txt
    def delete_file(self):
        file = request.args.get('file')
        LOG.info('The delete file is : %s', file)

        abs_path = os.path.join(SMB_FOLDER, file)
        os.remove(abs_path)

        return webob.Response(json={"accept": True})


def setup_app():
    agent_server = AgentServrer()

    return agent_server.app


app = setup_app()
serving.run_simple('0.0.0.0', 5000, app, threaded=True)
