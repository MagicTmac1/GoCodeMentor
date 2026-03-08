
function toggleClassDropdown() {
    const dropdown = document.getElementById('classDropdown');
    dropdown.style.display = dropdown.style.display === 'none' ? 'block' : 'none';
}

function hideClassDropdown() {
    document.getElementById('classDropdown').style.display = 'none';
}

function showCreateModal() {
    document.getElementById('createModal').style.display = 'flex';
}

function hideCreateModal() {
    document.getElementById('createModal').style.display = 'none';
}

function showDeleteClassModal() {
    document.getElementById('deleteClassModal').style.display = 'flex';
    // You would typically load the classes into the select element here
}

function hideDeleteClassModal() {
    document.getElementById('deleteClassModal').style.display = 'none';
}

async function showCreateAssignmentModal() {
    const topicSelect = document.getElementById('aiTopic');
    topicSelect.innerHTML = '<option>正在加载知识点...</option>';
    document.getElementById('createAssignmentModal').style.display = 'flex';

    try {
        const response = await fetch('/api/wisdom-graph'); 
        if (!response.ok) {
            throw new Error(`网络错误: ${response.status}`);
        }
        const graphData = await response.json();

        topicSelect.innerHTML = ''; // 清空加载提示

        if (!graphData.nodes || graphData.nodes.length === 0) {
            topicSelect.innerHTML = '<option>没有可用的知识点</option>';
            return;
        }

        graphData.nodes.forEach(node => {
            const option = document.createElement('option');
            option.value = node.name; // Use name as value for AI generation
            option.textContent = node.name;
            topicSelect.appendChild(option);
        });
    } catch (error) {
        topicSelect.innerHTML = '<option>知识点加载失败</option>';
        console.error('Failed to load topics:', error);
    }
}

function hideCreateAssignmentModal() {
    document.getElementById('createAssignmentModal').style.display = 'none';
}

function hidePublishModal() {
    document.getElementById('publishModal').style.display = 'none';
}

function hideViewAssignmentModal() {
    document.getElementById('viewAssignmentModal').style.display = 'none';
}

async function loadClasses() {
    console.log('loadClasses called');
    const classList = document.getElementById('classList');
    classList.innerHTML = '<div class="empty-state"><p>正在加载班级...</p></div>';

    try {
        const response = await fetch('/api/classes');
        if (!response.ok) {
            throw new Error(`网络错误: ${response.status}`);
        }
        const classes = await response.json();
        
        classList.innerHTML = ''; // 清空加载提示

        if (!classes || classes.length === 0) {
            const emptyState = document.createElement('div');
            emptyState.className = 'empty-state';
            emptyState.innerHTML = '<div class="icon">📚</div><h3>还没有创建班级</h3><p>点击右上角按钮创建你的第一个班级</p>';
            classList.appendChild(emptyState);
            return;
        }

        for (const c of classes) {
            const card = document.createElement('div');
            card.className = 'class-card';
            card.onclick = () => location.href = `/class/${c.ID}/students`;

            const title = document.createElement('h3');
            title.textContent = c.Name;

            const code = document.createElement('div');
            code.className = 'code';
            code.textContent = `班级码: ${c.Code}`;

            const stats = document.createElement('div');
            stats.className = 'stats';

            const studentStat = document.createElement('div');
            studentStat.innerHTML = `<div class="num" id="students-${c.ID}">...</div><div class="label">学生数</div>`;

            const assignmentStat = document.createElement('div');
            assignmentStat.innerHTML = `<div class="num" id="assignments-${c.ID}">...</div><div class="label">作业数</div>`;
            
            const unsubmittedStat = document.createElement('div');
            unsubmittedStat.innerHTML = `<div class="num" id="unsubmitted-${c.ID}">...</div><div class="label">未批改</div>`;

            const analysisBtn = document.createElement('div');
            analysisBtn.className = 'ai-analysis-btn';
            analysisBtn.innerHTML = '<span class="ai-badge">AI</span> 学情分析';
            analysisBtn.onclick = (e) => {
                e.stopPropagation();
                showStudentAnalysis(c.ID);
            };

            stats.append(studentStat, assignmentStat, unsubmittedStat);
            card.append(title, code, stats, analysisBtn);
            classList.appendChild(card);

            // Fetch stats for each class
            fetch(`/api/classes/${c.ID}/stats`).then(res => res.json()).then(data => {
                document.getElementById(`students-${c.ID}`).textContent = data.student_count;
                document.getElementById(`assignments-${c.ID}`).textContent = data.assignment_count;
                document.getElementById(`unsubmitted-${c.ID}`).textContent = data.unsubmitted_count;
            }).catch(err => console.error(`Failed to load stats for class ${c.ID}:`, err));
        }
    } catch (error) {
        classList.innerHTML = '<p style="color:red; text-align:center;">加载班级失败，请检查网络或稍后重试。</p>';
        console.error('Failed to load classes:', error);
    }
}

async function loadAssignments() {
    console.log('loadAssignments called');
    const assignmentList = document.getElementById('assignmentList');
    assignmentList.innerHTML = '<div class="empty-state"><p>正在加载作业...</p></div>';

    try {
        const response = await fetch('/api/assignments');
        if (!response.ok) {
            throw new Error(`网络错误: ${response.status}`);
        }
        const assignments = await response.json();
        
        assignmentList.innerHTML = ''; // 清空加载提示

        if (!assignments || assignments.length === 0) {
            const emptyState = document.createElement('div');
            emptyState.className = 'empty-state';
            emptyState.innerHTML = '<div class="icon">📝</div><h3>作业库为空</h3><p>点击右上角按钮创建作业</p>';
            assignmentList.appendChild(emptyState);
            return;
        }

        for (const a of assignments) {
            const card = document.createElement('div');
            card.className = 'class-card';

            const statusBadge = document.createElement('div');
            statusBadge.id = `status-${a.ID}`;
            statusBadge.className = 'assignment-status-badge draft';
            statusBadge.textContent = '草稿';

            const title = document.createElement('h3');
            title.textContent = a.Title;

            const actions = document.createElement('div');
            actions.className = 'assignment-actions';

            const viewButton = document.createElement('button');
            viewButton.className = 'btn btn-secondary';
            viewButton.textContent = '查看';
            viewButton.onclick = () => viewAssignmentDetail(a.ID);

            const publishButton = document.createElement('button');
            publishButton.className = 'btn';
            publishButton.textContent = '发布';
            publishButton.onclick = () => publishAssignment(a.ID);
            
            const deleteButton = document.createElement('button');
            deleteButton.className = 'btn btn-danger';
            deleteButton.innerHTML = '🗑️';
            deleteButton.onclick = () => deleteAssignment(a.ID);

            actions.append(viewButton, publishButton, deleteButton);
            card.append(statusBadge, title, actions);
            assignmentList.appendChild(card);

            // Fetch published status
            fetch(`/api/assignments/${a.ID}/published`).then(res => res.json()).then(publishedClasses => {
                if (publishedClasses && publishedClasses.length > 0) {
                    const badge = document.getElementById(`status-${a.ID}`);
                    badge.textContent = '已发布';
                    badge.className = 'assignment-status-badge published';
                }
            }).catch(err => console.error(`Failed to load published status for assignment ${a.ID}:`, err));
        }
    } catch (error) {
        assignmentList.innerHTML = '<p style="color:red; text-align:center;">加载作业失败，请检查网络或稍后重试。</p>';
        console.error('Failed to load assignments:', error);
    }
}

function createClass() {
    // Logic to create a class will be implemented here
    console.log("createClass function called");
}

function handleClassExcelUpload(input) {
    // Logic to handle Excel upload will be implemented here
    console.log("handleClassExcelUpload function called");
}

function deleteClass() {
    // Logic to delete a class will be implemented here
    console.log("deleteClass function called");
}

async function generateByAI() {
    const topic = document.getElementById('aiTopic').value;
    const difficulty = document.getElementById('aiDifficulty').value;
    const btn = document.getElementById('aiGenBtn');

    if (!topic) {
        alert('请选择一个知识点！');
        return;
    }

    btn.disabled = true;
    btn.innerHTML = '🧠 AI 正在出题中...';

    try {
        const response = await fetch('/api/assignments/generate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ topic, difficulty }),
        });

        const result = await response.json();

        if (!response.ok) {
            throw new Error(result.error || 'AI 作业生成失败');
        }

        hideCreateAssignmentModal();
        loadAssignments(); // Refresh the list to show the new assignment
        alert('AI 作业生成成功！');

    } catch (error) {
        alert(`生成失败: ${error.message}`);
    } finally {
        btn.disabled = false;
        btn.innerHTML = '🚀 开始 AI 生成';
    }
}

async function deleteAssignment(assignmentId) {
    if (!confirm('你确定要删除这个作业吗？此操作不可撤销。')) {
        return;
    }

    try {
        const response = await fetch(`/api/assignments/${assignmentId}`, { method: 'DELETE' });
        if (!response.ok) {
            const result = await response.json();
            throw new Error(result.error || '删除失败');
        }
        
        // Refresh the assignment list
        loadAssignments();

    } catch (error) {
        alert(`删除失败: ${error.message}`);
        console.error('Failed to delete assignment:', error);
    }
}

Object.assign(window.App, {
    // ... (existing functions)
    loadClasses,
    showCreateModal,
    hideCreateModal,
    createClass,
    showDeleteClassModal,
    hideDeleteClassModal,
    deleteClass,
    loadAssignments,
    showCreateAssignmentModal,
    hideCreateAssignmentModal,
    generateByAI,
    publishAssignment,
    hidePublishModal,
    confirmMultiPublish,
    viewAssignmentDetail,
    hideViewAssignmentModal,
    deleteAssignment,
    showStudentAnalysis,
    // ... (other functions)
});

async function publishAssignment(assignmentId) {
    currentPublishingAssignmentId = assignmentId;
    const modal = document.getElementById('publishModal');
    const classListEl = document.getElementById('classListForPublish');
    modal.style.display = 'flex';
    classListEl.innerHTML = '<p>正在加载班级列表...</p>';

    try {
        // Fetch all classes for the teacher
        const classResponse = await fetch('/api/classes');
        if (!classResponse.ok) throw new Error(`无法加载班级列表: ${classResponse.status}`);
        const classes = await classResponse.json();

        // Fetch already published classes for this assignment
        const publishedResponse = await fetch(`/api/assignments/${assignmentId}/published`);
        if (!publishedResponse.ok) throw new Error(`无法加载已发布状态: ${publishedResponse.status}`);
        const publishedClasses = await publishedResponse.json();
        const publishedClassIds = new Set(publishedClasses.map(pc => pc.ClassID));

        classListEl.innerHTML = '';

        if (!classes || classes.length === 0) {
            classListEl.innerHTML = '<p>你还没有创建任何班级。</p>';
            return;
        }

        classes.forEach(cls => {
            const isPublished = publishedClassIds.has(cls.ID);
            const item = document.createElement('div');
            item.className = `class-publish-item ${isPublished ? 'published' : ''}`;
            item.innerHTML = `
                <div class="class-info">
                    <input type="checkbox" id="class-checkbox-${cls.ID}" data-class-id="${cls.ID}" class="publish-checkbox" ${isPublished ? 'disabled' : ''}>
                    <label for="class-checkbox-${cls.ID}" class="class-name">${cls.Name}</label>
                    <span class="student-count">${cls.StudentCount || 0} 名学生</span>
                </div>
                <input type="date" class="deadline-input" id="deadline-${cls.ID}" ${isPublished ? 'disabled' : ''}>
                ${isPublished ? '<span class="status-tag published">已发放</span>' : '<span class="status-tag not-published">未发放</span>'}
            `;
            classListEl.appendChild(item);
        });

    } catch (error) {
        classListEl.innerHTML = `<p style="color:red;">${error.message}</p>`;
        console.error('Failed to prepare publish modal:', error);
    }
}

async function confirmMultiPublish() {
    if (!currentPublishingAssignmentId) {
        console.error("发布失败：未设置当前作业ID。");
        return;
    }

    const selectedClasses = [];
    const checkboxes = document.querySelectorAll('.publish-checkbox:checked');

    if (checkboxes.length === 0) {
        alert('请至少选择一个班级进行发布。');
        return;
    }

    let allDeadlinesSet = true;
    for (const cb of checkboxes) {
        const classId = cb.dataset.classId;
        const deadlineInput = document.getElementById(`deadline-${classId}`);
        if (deadlineInput && deadlineInput.value) {
            selectedClasses.push({ classId, deadline: deadlineInput.value });
        } else {
            const label = document.querySelector(`label[for='${cb.id}']`);
            const className = label ? label.textContent : `ID ${classId}`;
            alert(`发布失败：请为班级 "${className}" 设置一个截止日期。`);
            allDeadlinesSet = false;
            break; // 发现第一个错误就停止
        }
    }

    if (!allDeadlinesSet) {
        return;
    }

    const btn = document.querySelector('#publishModal .modal-buttons .btn');
    btn.disabled = true;
    btn.textContent = '发布中...';

    let successCount = 0;
    let errorCount = 0;

    for (const { classId, deadline } of selectedClasses) {
        try {
            const response = await fetch(`/api/assignments/${currentPublishingAssignmentId}/publish`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ class_id: classId, deadline: deadline }),
            });
            const result = await response.json();
            if (!response.ok) {
                throw new Error(result.error || '未知错误');
            }
            successCount++;
        } catch (error) {
            console.error(`发布到班级 ${classId} 失败:`, error);
            errorCount++;
            alert(`发布到班级 ${classId} 失败: ${error.message}`);
        }
    }

    btn.disabled = false;
    btn.textContent = '确认发布';

    alert(`发布操作完成！\n成功: ${successCount} 个班级\n失败: ${errorCount} 个班级`);

    // 如果全部成功，则关闭并刷新。否则，保留弹窗让用户修正。
    if (errorCount === 0 && successCount > 0) {
        hidePublishModal();
        loadAssignments(); 
    } else {
        // 如果有错误，刷新弹窗以显示哪些已成功，哪些仍可尝试
        publishAssignment(currentPublishingAssignmentId);
    }
}

async function viewAssignmentDetail(assignmentId) {
    const modal = document.getElementById('viewAssignmentModal');
    const titleEl = document.getElementById('viewAssignmentTitle');
    const contentEl = document.getElementById('viewAssignmentContent');

    modal.style.display = 'flex';
    titleEl.textContent = '正在加载...';
    contentEl.innerHTML = '';

    try {
        const response = await fetch(`/api/assignments/${assignmentId}`);
        if (!response.ok) {
            throw new Error(`网络错误: ${response.status}`);
        }
        const data = await response.json();

        titleEl.textContent = data.assignment.Title;
        
        let contentHtml = '';
        if (data.assignment.Description) {
            contentHtml += `<div class="assignment-description markdown-body">${marked.parse(data.assignment.Description)}</div>`;
        }

        if (data.questions && data.questions.length > 0) {
            contentHtml += '<div class="questions-container">';
            data.questions.forEach((q, index) => {
                const debug_html = `<pre style="background: #f0f0f0; border: 1px solid #ddd; padding: 10px; margin-top: 10px; font-size: 12px; white-space: pre-wrap; word-break: break-all;"><strong>DEBUG INFO:</strong>\nType: ${q.Type}\nOptions Raw: ${q.Options}\n</pre>`;

                let options_html = '';
                if (q.Type === 'choice' && q.Options) {
                    options_html += '<div class="options-container">';
                    try {
                        const optionsArray = JSON.parse(q.Options);
                        if (Array.isArray(optionsArray)) {
                            optionsArray.forEach((optionText, index) => {
                                const optionLabel = String.fromCharCode(65 + index); // A, B, C...
                                options_html += `<p class="option-item"><strong>${optionLabel}.</strong> ${optionText}</p>`;
                            });
                        } else {
                             console.error('Parsed options is not an array:', optionsArray);
                        }
                    } catch (e) {
                        console.error('Error parsing options:', q.Options, e);
                        options_html += `<p style="color:red;">Failed to load options.</p>`
                    }
                    options_html += '</div>';
                }

                let answer_html;
                if (q.Type === 'programming' || (q.Answer && q.Answer.includes('package main') && !q.Answer.includes('```'))) {
                    const formattedCode = marked.parse('```go\n' + q.Answer + '\n```');
                    answer_html = `<div class="answer-content"><strong>答案:</strong>${formattedCode}</div>`;
                } else {
                    answer_html = `<div class="answer-content-inline"><strong>答案:</strong> <span>${q.Answer || ''}</span></div>`;
                }
                contentHtml += `
                    <div class="question-item">
                        <div class="question-title"><strong>题目 ${index + 1}:</strong> ${q.Content}</div>
                        ${options_html}
                        ${answer_html}

                    </div>
                `;
            });
            contentHtml += '</div>';
        }
        contentEl.innerHTML = contentHtml;

        // 对新加载的内容中的代码块进行高亮
        contentEl.querySelectorAll('pre code').forEach((block) => {
            hljs.highlightBlock(block);
        });

    } catch (error) {
        titleEl.textContent = '加载失败';
        contentEl.innerHTML = `<p class="error-message">${error.message}</p>`;
        console.error('Failed to load assignment detail:', error);
    }
}

async function showStudentAnalysis(classId) {
    const modal = document.getElementById('analysisModal');
    const modalBody = document.getElementById('analysisModalBody');
    modal.style.display = 'flex';
    modalBody.innerHTML = `
        <div class="ai-analysis-loader">
            <div class="loader-icon">🧠</div>
            <h3>AI 正在分析中...</h3>
            <p>请稍候，我们正在为您生成学情报告</p>
        </div>
    `;

    try {
        const response = await fetch(`/api/classes/${classId}/ai-analysis`);
        if (!response.ok) {
            const errData = await response.json();
            throw new Error(errData.error || `分析失败: ${response.status}`);
        }
        const data = await response.json();

        if (data.report) {
            modalBody.innerHTML = `<div class="markdown-body">${marked.parse(data.report)}</div>`;
        } else {
            throw new Error('AI 返回的报告为空。');
        }

    } catch (error) {
        modalBody.innerHTML = `<p style="color:red; text-align: center;">${error.message}</p>`;
        console.error('Failed to get student analysis:', error);
    }
}

function submitNewResource() {
    console.log("submitNewResource function called");
}

function hideAddResourceModal() {
    console.log("hideAddResourceModal function called");
}

document.addEventListener('DOMContentLoaded', function () {
    const userData = {
        id: sessionStorage.getItem('user_id'),
        name: sessionStorage.getItem('user_name'),
        role: sessionStorage.getItem('user_role')
    };
    if (window.App) {
        window.App.init(userData);
        window.App.switchTab('classes'); // Default tab for teachers
    }
});
