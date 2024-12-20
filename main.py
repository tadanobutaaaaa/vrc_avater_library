import os
import re
import multiprocessing
import uvicorn
import signal
from pathlib import Path
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from typing import List
from googleSearch import google_search
from searchIcons import makeUnitypackageFile, settingFolderIcon

app = FastAPI()

origins = [
    'http://wails.localhost:34115',
    'http://wails.localhost',
]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

if not os.path.isdir('Images'):
    os.mkdir('Images')
if not os.path.isdir('Avaters'):
    os.mkdir('Avaters')

app.mount("/Images", StaticFiles(directory="Images"), name="Images")

@app.post("/image/get")
async def getImage():
    folderInformationList = []
    homeDirectory = os.path.expanduser("~")
    searchFolderPath = Path(homeDirectory) / "Downloads" # todo フロントエンドの設定画面などでどのフォルダを検索する対象にするか選べるようにする
    filePathList = list(searchFolderPath.rglob("*.unitypackage"))
    pattern = r"(_ver|ver|_|-v|_v)\d+(\.\d+)*|\.unitypackage$"

    for index, path in enumerate(filePathList):
        if index + 1 == 100: break
        nextPath = str(path).split(os.sep)
        cleaned = re.sub(pattern, "", os.path.basename(str(path)))
        cleanedPath = cleaned.replace("_"," ")
        subdirname = os.path.join(str(searchFolderPath), nextPath[nextPath.index("Downloads") + 1])
        folderInformationList.append({'fullPath': str(path), 'path': cleaned, 'subPath': cleanedPath,'subdirname': subdirname})

    returnInformationList = google_search(folderInformationList)
    
    existedPaths = []
    notExistedPaths = []
    
    for object in returnInformationList:
        if 'url' in object:
            existedPaths.append(object)
        else:
            notExistedPaths.append(object)
    return {
        "existedPaths":
            existedPaths
        ,
        "notExistedPaths":
            notExistedPaths
        ,
    }

@app.post("/thumbnail")
async def settingThumbnail(ThumbnailList: List[dict]):
    print(ThumbnailList)
    imagefile = './Images/'
    for file in ThumbnailList:
        settingFolderIcon(os.path.join(imagefile, file['path'] + '.png'), file['path'], file['subdirname'], makeUnitypackageFile(file['subdirname']))
    return{"status": "Success"}

@app.post("/shutdown")
async def shutdown():
    os.kill(os.getpid(), signal.SIGINT)
    return {"message": "正常にFastAPIが停止されました"}

if __name__ == '__main__':
    multiprocessing.freeze_support()
    uvicorn.run(app, port=8000, reload=False, workers=1)