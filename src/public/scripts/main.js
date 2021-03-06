document.addEventListener('DOMContentLoaded', () => {
    const mainArea = document.getElementById('main-area');
    const hostIpSpan = document.getElementById('host-ip');
    const filePickerBtn = document.getElementById('file-picker-btn');
    const clipboardCollection = document.getElementById('clipboard-collection');
    const clipboardSendBtn = document.getElementById('clipboard-add-btn');
    const clipboardContent = document.getElementById('clipboard-content');
    const deviceName = window.prompt('Please enter an unique device name', '');
    const fileTable = document.getElementById('file-table');
    const toggleIcon = document.getElementById('toggle-icon');
    const prevents = (e) => e.preventDefault();
    const modals = document.querySelectorAll('.modal');
    M.Modal.init(modals, null);
    const clipboard = new ClipboardJS('.copy');
    let maxFileSize;
    let socket;
    let mode = 'dark';

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
            mainArea.style.backgroundColor = 'unset';
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

    if ((window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) || mode === 'dark') {
        mode = 'dark';
        changeMode('dark');
        if(toggleIcon.textContent === 'brightness_high') {
            toggleIcon.textContent = 'brightness_4';
        }
    }

    document.getElementById('mode-toggle').addEventListener('click', e => {
        if (mode === 'dark') {
            mode = 'light';
            document.querySelector('body').classList.remove('dark');
        } else {
            mode = 'dark';
            document.querySelector('body').classList.add('dark')
        }

        if(toggleIcon.textContent === 'brightness_4') {
            toggleIcon.textContent = 'brightness_high';
        } else {
            toggleIcon.textContent = 'brightness_4';
        }
    });

    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', e => {
        mode = e.matches ? 'dark' : 'light';
        changeMode(mode);
        if(mode === 'light') {
            toggleIcon.textContent = 'brightness_high';
        } else {
            toggleIcon.textContent = 'brightness_4';
        }
    });

    function processFiles(files) {
        files.forEach(file => {
            if (file.size > convertTo(maxFileSize.replace('MB', '').trim(), 'M', 'B')) {
                alert('File size of ' + file.name + ' exceed allowed size ' + maxFileSize);
                return;
            }
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
                maxFileSize = response.data.MaxFileSize;
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

    function changeMode(mode) {
        if (mode === 'dark') {
            document.querySelector('body').classList.add('dark'); 
        } else {
            document.querySelector('body').classList.remove('dark');
        }
    }

    getMeta();
});