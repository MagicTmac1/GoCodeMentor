
function goToChat() {
    window.location.href = `/student/${window.App.user.id}/chats?name=${encodeURIComponent(window.App.user.name || '学生')}`;
}

async function loadAssignments() {
    const listEl = document.getElementById('assignmentList');
    listEl.innerHTML = '<div class="empty-state"><p>正在加载作业...</p></div>';

    try {
        const res = await fetch(`/api/student/assignments`);
        if (!res.ok) {
            throw new Error(`网络错误: ${res.status}`);
        }
        const data = await res.json();
        
        listEl.innerHTML = ''; // 清空加载提示

        if (!data.assignments || data.assignments.length === 0) {
            listEl.innerHTML = '<div class="empty-state"><div class="icon">📝</div><p>暂无作业</p></div>';
            return;
        }

        data.assignments.forEach(item => {
            const assign = item.assignment;
            const sub = item.submission;
            const statusText = item.status;

            let statusClass = 'unsubmitted';
            if (statusText === '已批改' || statusText === '已查看') statusClass = 'graded';
            else if (statusText === '已提交') statusClass = 'submitted';
            
            // Defensively handle null or undefined score
            const score = (sub && sub.total_score !== null && sub.total_score !== undefined) ? sub.total_score : '--';
            
            const card = document.createElement('div');
            // 复用 teacher.js 中的 class-card 样式
            card.className = 'class-card assignment-card-student'; 

            const statusBadge = document.createElement('div');
            statusBadge.className = `assignment-status-badge ${statusClass}`;
            statusBadge.textContent = statusText;

            const title = document.createElement('h3');
            title.textContent = assign.Title || '未命名作业';

            const deadline = document.createElement('div');
            deadline.className = 'code'; // 复用班级码的样式
            deadline.textContent = `截止: ${assign.Deadline ? new Date(assign.Deadline).toLocaleString() : '无'}`;
            deadline.style.marginBottom = '20px';

            const stats = document.createElement('div');
            stats.className = 'stats';
            stats.style.justifyContent = 'flex-start'; // 左对齐

            const scoreStat = document.createElement('div');
            scoreStat.innerHTML = `<div class="num">${score}</div><div class="label">得分</div>`;

            const actionButtonContainer = document.createElement('div');
            actionButtonContainer.className = 'assignment-actions'; // 复用样式
            actionButtonContainer.style.position = 'absolute';
            actionButtonContainer.style.right = '25px';
            actionButtonContainer.style.bottom = '25px';

            if (statusClass === 'unsubmitted') {
                const link = document.createElement('a');
                link.href = `/assignments/do?id=${assign.ID}`;
                link.className = 'btn'; // 使用主按钮样式
                link.textContent = '去完成';
                actionButtonContainer.appendChild(link);
            } else {
                const button = document.createElement('button');
                button.className = 'btn btn-secondary'; // 使用次要按钮样式
                button.textContent = '查看详情';
                button.onclick = () => viewAssignmentDetail(assign.ID);
                actionButtonContainer.appendChild(button);
            }
            
            stats.append(scoreStat);
            card.append(statusBadge, title, deadline, stats, actionButtonContainer);
            listEl.appendChild(card);
        });
    } catch (e) {
        listEl.innerHTML = '<p style="color:red; text-align:center;">加载作业失败，请检查网络或稍后重试。</p>';
        console.error('Failed to load assignments:', e);
    }
}

async function viewAssignmentDetail(id) {
    const modal = document.getElementById('assignmentModal');
    const body = document.getElementById('modalBody');
    modal.style.display = 'flex';
    body.innerHTML = '<p>正在加载详情...</p>';

    try {
        const res = await fetch(`/api/assignments/${id}/student/${window.App.user.id}`);
        if (!res.ok) {
            throw new Error(`网络错误: ${res.status}`);
        }
        const data = await res.json();
        const sub = data.submission || data.Submission || {};
        const parsedQScores = (sub && sub.question_scores) ? sub.question_scores : {};
        const parsedQFeedback = (sub && sub.question_feedback) ? sub.question_feedback : {};
        const parsedStudentAnswers = (sub && sub.answers) ? sub.answers : {};
        
        body.innerHTML = ''; // Clear loading text

        const title = document.createElement('h2');
        title.textContent = data.assignment.Title;
        title.style.marginBottom = '20px';
        title.style.color = '#1f2937';
        body.appendChild(title);

        const summaryContainer = document.createElement('div');
        summaryContainer.style.background = '#f8f9fa';
        summaryContainer.style.padding = '20px';
        summaryContainer.style.borderRadius = '12px';
        summaryContainer.style.marginBottom = '25px';
        summaryContainer.style.border = '1px solid #eef2ff';

        const summaryFlex = document.createElement('div');
        summaryFlex.style.display = 'flex';
        summaryFlex.style.gap = '30px';
        summaryFlex.style.marginBottom = '15px';

        const statusP = document.createElement('p');
        statusP.innerHTML = `<strong>状态:</strong> ${sub.status === 'graded' ? '<span style="color: #059669; font-weight: 600;">✅ 已批改</span>' : '<span style="color: #2563eb; font-weight: 600;">📤 已提交</span>'}`;

        const scoreP = document.createElement('p');
        scoreP.innerHTML = `<strong>总分:</strong> <span style="font-size: 20px; color: #4f46e5; font-weight: 700;">${sub.total_score || '--'}</span>`;

        summaryFlex.append(statusP, scoreP);
        summaryContainer.appendChild(summaryFlex);

        if (sub.ai_feedback) {
            const aiFeedbackContainer = document.createElement('div');
            aiFeedbackContainer.style.marginTop = '20px';
            
            const aiFeedbackTitle = document.createElement('strong');
            aiFeedbackTitle.textContent = '🤖 AI 评改报告:';
            aiFeedbackTitle.style.display = 'block';
            aiFeedbackTitle.style.marginBottom = '12px';
            aiFeedbackTitle.style.color = '#374151';
            aiFeedbackTitle.style.fontSize = '16px';
            aiFeedbackContainer.appendChild(aiFeedbackTitle);

            const aiFeedbackBody = document.createElement('div');
            aiFeedbackBody.className = 'markdown-body';
            aiFeedbackBody.style.padding = '20px';
            aiFeedbackBody.style.background = '#fff';
            aiFeedbackBody.style.borderRadius = '12px';
            aiFeedbackBody.style.border = '1px solid #e5e7eb';
            aiFeedbackBody.style.fontSize = '14px';
            aiFeedbackBody.style.lineHeight = '1.6';
            aiFeedbackBody.style.color = '#374151';
            // WARNING: Using innerHTML with markdown. Ensure marked is configured with a sanitizer.
            aiFeedbackBody.innerHTML = marked.parse(sub.ai_feedback);
            aiFeedbackContainer.appendChild(aiFeedbackBody);

            summaryContainer.appendChild(aiFeedbackContainer);
        }

        body.appendChild(summaryContainer);

        if (data.questions && data.questions.length > 0) {
            const questionsTitle = document.createElement('h3');
            questionsTitle.textContent = '答题详情';
            questionsTitle.style.margin = '25px 0 15px';
            questionsTitle.style.color = '#1f2937';
            body.appendChild(questionsTitle);

            data.questions.forEach((q, idx) => {


                const studentAns = parsedStudentAnswers[q.ID] || '';
                const qScore = parsedQScores[q.ID] !== undefined ? parsedQScores[q.ID] : null;
                const qFeedback = parsedQFeedback[q.ID] || '';

                const questionContainer = document.createElement('div');
                questionContainer.style.margin = '15px 0';
                questionContainer.style.padding = '18px';
                questionContainer.style.border = '1px solid #eee';
                questionContainer.style.borderRadius = '12px';
                questionContainer.style.background = '#fff';
                questionContainer.style.boxShadow = '0 2px 4px rgba(0,0,0,0.02)';

                // --- New logic to handle choice question display --- //
                let studentAnsText = studentAns;
                let correctAnswerText = q.Answer;

                if (q.Type === 'choice' && q.Options) {
                    try {
                        const optionsArray = JSON.parse(q.Options);

                        // Convert student's answer (e.g., 'A') to full option text
                        if (studentAns && studentAns.match(/^[A-Z]$/)) {
                            const studentOptionIndex = studentAns.charCodeAt(0) - 'A'.charCodeAt(0);
                            if (studentOptionIndex >= 0 && studentOptionIndex < optionsArray.length) {
                                studentAnsText = optionsArray[studentOptionIndex];
                            }
                        }

                        // Convert correct answer (e.g., 'A' or '1') to full option text
                        if (correctAnswerText && correctAnswerText.match(/^[A-Z]$/)) {
                            const correctOptionIndex = correctAnswerText.charCodeAt(0) - 'A'.charCodeAt(0);
                            if (correctOptionIndex >= 0 && correctOptionIndex < optionsArray.length) {
                                correctAnswerText = optionsArray[correctOptionIndex];
                            }
                        } else if (correctAnswerText && !isNaN(parseInt(correctAnswerText))) { // Handle index-based answer like "1"
                            const correctOptionIndex = parseInt(correctAnswerText) - 1;
                             if (correctOptionIndex >= 0 && correctOptionIndex < optionsArray.length) {
                                correctAnswerText = optionsArray[correctOptionIndex];
                            }
                        }
                    } catch (e) {
                        console.error('Error parsing question options for display:', e);
                    }
                }
                // --- End of new logic --- //

                const questionTitleP = document.createElement('p');
                questionTitleP.style.fontWeight = '600';
                questionTitleP.style.marginBottom = '12px';
                questionTitleP.innerHTML = `<strong>${idx + 1}. ${q.Content}</strong> <span style="font-weight: 400; color: #667eea;">(${q.Score}分)</span>`;

                const studentAnswerP = document.createElement('p');
                studentAnswerP.style.marginBottom = '8px';
                studentAnswerP.innerHTML = `<strong>你的回答:</strong> <span style="color: #333;">${studentAnsText}</span>`;

                const correctAnswerP = document.createElement('p');
                correctAnswerP.style.marginBottom = '12px';

                // 仅在已批改状态下显示正确答案
                if (sub.status === 'graded') {
                    if (q.Type === 'code') {
                        correctAnswerP.innerHTML = '<strong>参考答案:</strong>';
                        const pre = document.createElement('pre');
                        pre.style.background = '#f8f9fa';
                        pre.style.padding = '12px';
                        pre.style.borderRadius = '8px';
                        pre.style.whiteSpace = 'pre-wrap';
                        pre.style.wordWrap = 'break-word';
                        pre.style.fontSize = '13px';
                        
                        const code = document.createElement('code');
                        code.textContent = correctAnswerText;
                        
                        pre.appendChild(code);
                        correctAnswerP.appendChild(pre);
                    } else {
                        correctAnswerP.innerHTML = `<strong>参考答案:</strong> <span style="color: #28a745; font-weight: 500;">${correctAnswerText}</span>`;
                    }
                }

                const scoreFeedbackContainer = document.createElement('div');
                if (qScore !== null || qFeedback) {
                    scoreFeedbackContainer.style.background = '#f8f9fa';
                    scoreFeedbackContainer.style.padding = '12px';
                    scoreFeedbackContainer.style.borderRadius = '8px';
                    scoreFeedbackContainer.style.marginTop = '12px';
                    scoreFeedbackContainer.style.borderLeft = '4px solid #667eea';

                    if (qScore !== null) {
                        const scoreResultP = document.createElement('p');
                        scoreResultP.innerHTML = `<strong>得分:</strong> <span style="font-weight: 700; font-size: 16px; color: ${qScore >= (q.Score / 2) ? '#28a745' : '#dc3545'};">${qScore}</span>`;
                        scoreFeedbackContainer.appendChild(scoreResultP);
                    }
                    if (qFeedback) {
                        const feedbackP = document.createElement('p');
                        feedbackP.style.marginTop = '8px';
                        feedbackP.innerHTML = `<strong>教师评语:</strong> ${qFeedback}`;
                        scoreFeedbackContainer.appendChild(feedbackP);
                    }
                }

                questionContainer.append(questionTitleP, studentAnswerP);

                if (sub.status === 'graded') {
                    questionContainer.appendChild(correctAnswerP);
                }

                if (qScore !== null || qFeedback) {
                    questionContainer.appendChild(scoreFeedbackContainer);
                }



                body.appendChild(questionContainer);
            });
        } else if (sub.code) {
             const codeTitle = document.createElement('h3');
            codeTitle.textContent = '提交的代码';
            codeTitle.style.margin = '25px 0 15px';
            codeTitle.style.color = '#1f2937';
            body.appendChild(codeTitle);

            const codeContainer = document.createElement('div');
            codeContainer.style.position = 'relative';

            const pre = document.createElement('pre');
            pre.style.background = '#1e1e1e';
            pre.style.color = '#d4d4d4';
            pre.style.padding = '20px';
            pre.style.borderRadius = '12px';
            pre.style.overflow = 'auto';
            pre.style.whiteSpace = 'pre-wrap';
            pre.style.wordWrap = 'break-word';

            const code = document.createElement('code');   
            code.textContent = sub.code;

            pre.appendChild(code);
            codeContainer.appendChild(pre);
            body.appendChild(codeContainer);
        }

        body.querySelectorAll('pre code').forEach(el => hljs.highlightElement(el));
    } catch (e) {
        console.error('Error in viewAssignmentDetail:', e);
        body.innerHTML = '<p style="color:red; padding: 20px; text-align: center;">加载详情失败，请检查网络或稍后重试。</p>';
    }
}

function closeModal() {
    document.getElementById('assignmentModal').style.display = 'none';
}

function showAiSuggestions() {
    const suggestionsSection = document.getElementById('aiSuggestionsSection');
    suggestionsSection.style.display = 'block';
    const suggestionsContent = document.getElementById('aiSuggestionsContent');
    suggestionsContent.innerHTML = `
        <div style="padding: 20px; background-color: #f5f7fa; border-radius: 8px;">
            <h4 style="margin-bottom: 15px; color: #303133;">针对您的学习情况，AI建议如下：</h4>
            <p style="color: #606266; line-height: 1.8; font-size: 15px;">
                您在 <strong>「Go语言并发编程」</strong> 考点正确率较低，该考点是区块链分布式系统开发的核心基础。
                建议结合《区块链导论》中分布式节点通信知识点学习，同时完成题库中「goroutine实现区块链节点简单通信」专项题，为后续《分布式系统与存储》课程打好基础。
            </p>
        </div>
    `;
}

document.addEventListener('DOMContentLoaded', function () {
    const userData = {
        id: sessionStorage.getItem('user_id'),
        name: sessionStorage.getItem('user_name'),
        role: sessionStorage.getItem('user_role')
    };
    window.App.init(userData);

    if (!userData.id || userData.role !== 'student') {
        window.location.href = '/login';
        throw new Error('Redirecting to login...'); // 停止后续脚本执行
    }

    const urlParams = new URLSearchParams(window.location.search);
    const defaultTab = urlParams.get('tab') || 'course-details';
    window.App.switchTab(defaultTab);
});
