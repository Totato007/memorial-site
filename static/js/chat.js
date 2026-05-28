var lastMsgID = 0;

// 初始化
(function() {
    var msgs = document.querySelectorAll('.chat-msg');
    if (msgs.length > 0) {
        lastMsgID = parseInt(msgs[msgs.length - 1].dataset.id || '0');
    }
    poll();
})();

function poll() {
    var box = document.getElementById('chat-box');
    if (!box) return;
    var relID = box.dataset.rel || '0';
    var myUID = parseInt(box.dataset.user || '0');

    fetch('/chat/poll?rel_id=' + relID + '&last_id=' + lastMsgID)
        .then(function(r) { return r.json(); })
        .then(function(msgs) {
            if (msgs && msgs.length > 0) {
                msgs.forEach(function(m) {
                    // 跳过自己刚发的（已通过 sendMessage 即时显示）
                    if (m.id <= lastMsgID) return;
                    var mine = m.sender_id === myUID ? 'msg-mine' : 'msg-theirs';
                    appendMsg(box, m, mine);
                    lastMsgID = m.id;
                });
                box.scrollTop = box.scrollHeight;
            }
        })
        .catch(function() {})
        .finally(function() { setTimeout(poll, 4000); });
}

function sendMessage(relID) {
    var input = document.getElementById('chat-input');
    var fileInput = document.getElementById('chat-image');
    var content = input.value.trim();
    if (!content && !fileInput.files[0]) return false;

    var fd = new FormData();
    fd.append('rel_id', relID);
    fd.append('content', content);
    if (fileInput.files[0]) fd.append('image', fileInput.files[0]);

    var err = document.getElementById('chat-error');
    err.classList.add('hidden');

    // 禁用按钮防重复提交
    var btn = document.querySelector('#chat-form button[type=submit]');
    btn.disabled = true;

    fetch('/chat/send', {
        method: 'POST', body: fd,
        headers: {'X-Requested-With': 'XMLHttpRequest'}
    })
    .then(function(r) { return r.json(); })
    .then(function(d) {
        if (d.error) {
            err.textContent = d.error;
            err.classList.remove('hidden');
        } else {
            input.value = '';
            fileInput.value = '';
            document.getElementById('img-label').textContent = '图片';
            // 立即显示自己发的消息
            var box = document.getElementById('chat-box');
            var myUID = parseInt(box.dataset.user || '0');
            appendMsg(box, d, 'msg-mine');
            lastMsgID = d.id;
            box.scrollTop = box.scrollHeight;
        }
    })
    .catch(function() { err.textContent = '发送失败'; err.classList.remove('hidden'); })
    .finally(function() { btn.disabled = false; });

    return false;
}

function appendMsg(box, m, cls) {
    var div = document.createElement('div');
    div.className = 'chat-msg ' + cls;
    div.dataset.id = m.id;
    var html = '<div class="msg-sender">' + escHtml(m.sender ? m.sender.nickname : '') + '</div>';
    if (m.content) html += '<div class="msg-text">' + escHtml(m.content) + '</div>';
    if (m.image_path) html += '<div class="msg-image"><img src="/uploads/' + escHtml(m.image_path) + '" loading="lazy" onclick="this.classList.toggle(\'zoomed\')"></div>';
    var timeStr = (m.created_at || '').substring(5, 16).replace('T', ' ');
    html += '<div class="msg-time">' + timeStr + '</div>';
    div.innerHTML = html;
    box.appendChild(div);
}

function escHtml(s) {
    if (!s) return '';
    return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}
