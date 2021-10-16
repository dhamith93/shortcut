document.addEventListener('DOMContentLoaded', () => {
    const mainArea = document.getElementById('main-area');
    const hostIpSpan = document.getElementById('host-ip');
    const filePickerBtn = document.getElementById('file-picker-btn');
    const clipboardCollection = document.getElementById('clipboard-collection');
    const clipboardSendBtn = document.getElementById('clipboard-add-btn');
    const clipboardContent = document.getElementById('clipboard-content');
    const deviceName = window.prompt('Please enter an unique device name', '');
    const fileTable = document.getElementById('file-table');
    const prevents = (e) => e.preventDefault();
    const modals = document.querySelectorAll('.modal');
    M.Modal.init(modals, null);
    const clipboard = new ClipboardJS('.copy');
    let socket;

    if (deviceName && deviceName.length > 0) {
        document.getElementById('self-device-name').innerHTML = deviceName;
    } else {
        location.reload();
    }

    filePickerBtn.addEventListener('click', e => {
        const input = document.createElement('input');
        input.type = 'file';
        input.setAttribute('multiple', 'true');
        input.onchange = e => { 
            const files = e.target.files; 
            processFiles([...files]);
        }
        input.click();
    });

    clipboardSendBtn.addEventListener('click', e => {
        const content = clipboardContent.value;
        const data = {
            'DeviceName': deviceName,
            'Content': content
        };
        
        axios.post('clipboard', data).then((r) => {
            clipboardContent.value = '';
        }, (error) => {
            console.error(error);
            alert('Error adding the clipboard item... ' + error);
        }); 
    });

    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(evt => {
        mainArea.addEventListener(evt, prevents);
    });

    ['dragenter', 'dragover'].forEach(evt => {
        mainArea.addEventListener(evt, e => {
            mainArea.style.backgroundColor = '#DCDCDC';
        });
    });

    ['dragleave', 'drop'].forEach(evt => {
        mainArea.addEventListener(evt, e => {
            mainArea.style.backgroundColor = '#FFFFFF';
        });
    });

    mainArea.addEventListener('drop', e => {
        const files = e.dataTransfer.files;
        processFiles([...files]);
    });

    clipboard.on('success', function(e) {
        const span = document.getElementById(e.trigger.dataset.copiedLabel);
        if (span) {
            span.style.opacity = 1;
            setTimeout(() => {
                span.style.opacity = 0;
            }, 1500);
            e.clearSelection();
        }
    });

    function processFiles(files) {
        files.forEach(file => {
            const formData = new FormData();            
            formData.append('file', file);
            formData.append('device-name', deviceName);
            const config = {
                headers: {
                    'Content-Type': 'multipart/form-data'
                },
                onUploadProgress: (progressEvent) => {
                    document.getElementById('upload-progressbar').style.display = 'block';
                    let percentCompleted = Math.round( (progressEvent.loaded * 100) / progressEvent.total );
                    document.getElementById('progress-completed').style.width =  percentCompleted + '%';
                }
            };
            axios.post('upload', formData, config).then((r) => {}, (error) => {
                console.error(error);
                alert('Error uploading the file... ' + error);
            }).finally(() => {
                document.getElementById('upload-progressbar').style.display = 'none';
            });
        });
    }

    function getFileList() {
        axios.get('files').then((response) => {
            if (response.data) {
                populateTable(fileTable, response.data);
            }
        }, (error) => console.error(error));
    }

    function getClipboardItems() {
        axios.get('clipboard').then((response) => {
            if (response.data) {
                processClipboardItems(response.data);
            }
        }, (error) => console.error(error));
    }

    function processClipboardItems(data) {
        clearElement(clipboardCollection).then(() => {
            data.reverse();
            for (let i = 0; i < data.length; i++) {
                const item = data[i];
                clipboardCollection.innerHTML += buildCollectionItem(i, item.DeviceName, item.Content);
            }
        })
    }

    function getMeta() {
        axios.get('meta').then((response) => {
            if (response.data) {
                hostIpSpan.innerHTML = response.data.Url;
                socket = new WebSocket('ws://' + response.data.Url.replace('http://', '') + '/ws');
                socket.addEventListener('error', e => {
                    alert('Connection rejected');
                    document.getElementsByTagName('body')[0].style.display = 'none';
                });
                socket.addEventListener('message', e => {
                    try {
                        const data = JSON.parse(e.data);
                        data.forEach(item => {
                            switch (item.MsgType) {
                                case 'clipboardItems':
                                    processClipboardItems(item.Msg);
                                    break;
                                case 'fileList':
                                    populateTable(fileTable, item.Msg);
                                    break;
                                default:
                                    break;
                            }
                        });
                    } catch (e) {
                        // fall back to API if parsing JSON failed
                        getClipboardItems();
                        getFileList();
                    }
                });
            }
        }, (error) => {
            console.error(error);
        });
    }

    getMeta();
});