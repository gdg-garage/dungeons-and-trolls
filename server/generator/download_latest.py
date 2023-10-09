from github import Github
from github import Auth

import requests
import zipfile
import os
import stat

GENERATOR_REPO = "malytomas/dntGenerator"
ARCHIVE_PATH = "generator.zip"
API_KEY_PATH = "github_api_key"


def get_gh_api_key():
    with open(API_KEY_PATH, "r") as f:
        return f.read()

def download_archive(url: str):
    # pygithub does not provide method for this.
    with open(ARCHIVE_PATH, "wb") as f:
        r = requests.get(url, allow_redirects=True, headers={
            "Accept": "application/vnd.github+json",
            "Authorization" : f"Bearer {get_gh_api_key()}",
        })
        f.write(r.content)

def extract_binary():
    binary_name = "dntgenerator"
    with zipfile.ZipFile(ARCHIVE_PATH, mode="r") as archive:
        # print(archive.printdir())
        archive.extract(binary_name)
        archive.extract("libcage-core.so")
        os.chmod(binary_name, os.stat(binary_name).st_mode | stat.S_IEXEC)

def find_latest_artifact_url():
    auth = Auth.Token(get_gh_api_key())

    g = Github(auth=auth)
    repo = g.get_repo(GENERATOR_REPO)

    latest_valid_run = None
    for r in repo.get_workflow("deploy.yml").get_runs(branch="master", status="success"): #, exclude_pull_requests=True):
        latest_valid_run = r
        break

    for artifact in repo.get_artifacts():
        if artifact.workflow_run.id != latest_valid_run.id:
            continue
        return artifact.archive_download_url

def main():
    artifact_url  = find_latest_artifact_url()
    download_archive(artifact_url)
    extract_binary()

if __name__ == '__main__':
    main()