// 反馈管理系统 - 修复版本
// 适配后端API: /api/feedback (单数) 和直接数组响应

// ==================== 配置常量 ====================
const FEEDBACK_CONFIG = {
    API: {
        BASE: '/api',
        FEEDBACKS: '/api/feedback',  // 修正为单数，匹配后端路由
        FEEDBACK_LIKE: (id) => `/api/feedback/${id}/like`  // 修正为单数
    },
    UI: {
        ANIMATION_DURATION: 300,
        ITEMS_PER_PAGE: 10,
        MAX_TITLE_LENGTH: 100,
        MAX_CONTENT_LENGTH: 2000
    }
};

// ==================== 全局状态 ====================
const FeedbackState = {
    // 用户信息
    user: {
        id: '',
        role: 'student',
        name: ''
    },
    // 过滤条件
    filter: {
        type: '',
        status: '',
        search: ''
    },
    // 反馈列表
    feedbacks: [],
    // UI状态
    ui: {
        isLoading: false,
        showForm: false,
        showDetailModal: false,
        currentFeedbackId: null,
        currentPage: 1,
        totalPages: 1
    },
    // 统计信息
    stats: {
        total: 0,
        pending: 0,
        resolved: 0,
        myFeedbacks: 0
    }
};

// ==================== API 客户端 ====================
const FeedbackAPI = {
    async request(url, options = {}) {
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'same-origin'
        };
        
        try {
            const response = await fetch(url, { ...defaultOptions, ...options });
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || '请求失败');
            }
            
            return data;  // 直接返回数据，不包装
        } catch (error) {
            console.error('API请求失败:', error);
            throw error;
        }
    },

    // 获取反馈列表 - 后端直接返回数组
    async fetchFeedbacks(params = {}) {
        try {
            const queryParams = new URLSearchParams();
            
            // 只添加非空参数
            if (params.type) queryParams.append('type', params.type);
            if (params.status) queryParams.append('status', params.status);
            if (params.search) queryParams.append('search', params.search);
            
            const url = queryParams.toString() 
                ? `${FEEDBACK_CONFIG.API.FEEDBACKS}?${queryParams}`
                : FEEDBACK_CONFIG.API.FEEDBACKS;
                
            const feedbacks = await this.request(url, { method: 'GET' });
            
            // 后端直接返回反馈数组，不是包装对象
            return Array.isArray(feedbacks) ? feedbacks : [];
        } catch (error) {
            FeedbackUI.showNotification('获取反馈列表失败', 'error');
            throw error;
        }
    },

    // 创建反馈
    async createFeedback(data) {
        return await this.request(FEEDBACK_CONFIG.API.FEEDBACKS, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    // 点赞反馈
    async likeFeedback(id) {
        return await this.request(FEEDBACK_CONFIG.API.FEEDBACK_LIKE(id), {
            method: 'POST'
        });
    },

    // 获取单条反馈详情
    async fetchFeedbackDetail(id) {
        return await this.request(`/api/feedback/${id}`, {
            method: 'GET'
        });
    }
};

// ==================== UI 组件 ====================
const FeedbackUI = {
    // 渲染反馈列表
    renderFeedbacks(feedbacks) {
        const container = document.getElementById('feedback-list');
        if (!container) return;
        
        if (!feedbacks || feedbacks.length === 0) {
            container.innerHTML = '<div class="empty-state">暂无反馈</div>';
            return;
        }

        const pageSize = FEEDBACK_CONFIG.UI.ITEMS_PER_PAGE || 10;
        const totalPages = Math.max(1, Math.ceil(feedbacks.length / pageSize));
        FeedbackState.ui.totalPages = totalPages;

        // 保证当前页在有效范围内
        if (FeedbackState.ui.currentPage > totalPages) {
            FeedbackState.ui.currentPage = totalPages;
        }
        if (FeedbackState.ui.currentPage < 1) {
            FeedbackState.ui.currentPage = 1;
        }

        const start = (FeedbackState.ui.currentPage - 1) * pageSize;
        const pageItems = feedbacks.slice(start, start + pageSize);

        const listHtml = pageItems.map(f => this.createFeedbackCard(f)).join('');
        const paginationHtml = this.renderPagination();

        container.innerHTML = `
            <div class="feedback-list-body">
                ${listHtml}
            </div>
            ${paginationHtml}
        `;
    },

    // 创建单个反馈卡片
    createFeedbackCard(feedback) {
        // 处理日期格式
        let dateStr = '未知时间';
        if (feedback.CreatedAt) {
            try {
                const date = new Date(feedback.CreatedAt);
                dateStr = date.toLocaleString('zh-CN');
            } catch (e) {
                dateStr = feedback.CreatedAt;
            }
        }
        
        const isMyFeedback = feedback.AnonymousID === FeedbackState.user.id || 
                           (feedback.userId && feedback.userId === FeedbackState.user.id);
        const likeClass = feedback.Liked ? 'liked' : '';

        // 截断内容，只展示一小段摘要
        const rawContent = (feedback.Content || '').replace(/\s+/g, ' ').trim();
        const snippet = rawContent
            ? (rawContent.length > 120 ? rawContent.slice(0, 120) + '…' : rawContent)
            : '';

        const typeLabel = this.getTypeLabel(feedback.Type);
        const statusLabel = this.getStatusLabel(feedback.Status);

        // 使用更接近 GitHub Issues 的一行样式
        return `
            <div class="feedback-card feedback-row ${feedback.Status || 'pending'}" 
                 data-id="${feedback.ID}"
                 onclick="FeedbackController.viewFeedbackDetail(${feedback.ID})">
                <div class="feedback-row-main">
                    <div class="feedback-row-title">
                        <a href="javascript:void(0)" class="feedback-title-link">
                            ${feedback.Title || '无标题'}
                        </a>
                        <span class="feedback-label feedback-label--${feedback.Type || 'other'}">${typeLabel}</span>
                        <span class="feedback-status-pill feedback-status-pill--${feedback.Status || 'pending'}">${statusLabel}</span>
                        ${isMyFeedback ? '<span class="feedback-mine-pill">我的反馈</span>' : ''}
                    </div>
                    <div class="feedback-row-meta">
                        <span>${feedback.AnonymousID || '匿名用户'}</span>
                        <span>${dateStr}</span>
                        <span>👍 ${feedback.LikeCount || 0}</span>
                    </div>
                </div>
                <div class="feedback-row-actions">
                    <button class="feedback-like-chip ${likeClass}" 
                            onclick="event.stopPropagation(); FeedbackController.toggleLike(${feedback.ID})" 
                            title="点赞">
                        👍 ${feedback.LikeCount || 0}
                    </button>
                </div>
            </div>
        `;
    },

    // 分页渲染
    renderPagination() {
        const totalPages = FeedbackState.ui.totalPages || 1;
        if (totalPages <= 1) return '';

        const current = FeedbackState.ui.currentPage;
        const items = [];

        const pushPage = (page) => {
            const active = page === current;
            items.push(
                `<li class="${active ? 'active' : ''}">
                    ${active
                        ? `<span>${page}</span>`
                        : `<a href="javascript:void(0)" onclick="FeedbackController.goToPage(${page})">${page}</a>`}
                 </li>`
            );
        };

        // 上一页
        items.push(
            `<li class="${current === 1 ? 'disabled' : ''}">
                ${current === 1
                    ? '<span>«</span>'
                    : `<a href="javascript:void(0)" onclick="FeedbackController.goToPage(${current - 1})">«</a>`}
             </li>`
        );

        // 简单分页逻辑：最多展示 5 个页码
        const maxShown = 5;
        let start = Math.max(1, current - 2);
        let end = Math.min(totalPages, start + maxShown - 1);
        if (end - start + 1 < maxShown) {
            start = Math.max(1, end - maxShown + 1);
        }

        for (let p = start; p <= end; p++) {
            pushPage(p);
        }

        // 下一页
        items.push(
            `<li class="${current === totalPages ? 'disabled' : ''}">
                ${current === totalPages
                    ? '<span>»</span>'
                    : `<a href="javascript:void(0)" onclick="FeedbackController.goToPage(${current + 1})">»</a>`}
             </li>`
        );

        return `<ul class="pagination">${items.join('')}</ul>`;
    },

    // 更新统计卡片
    updateStats(feedbacks) {
        if (!Array.isArray(feedbacks)) return;
        
        const stats = {
            total: feedbacks.length,
            pending: feedbacks.filter(f => f.Status === 'open' || f.Status === 'pending').length,
            resolved: feedbacks.filter(f => f.Status === 'resolved' || f.Status === 'closed').length,
            myFeedbacks: feedbacks.filter(f => f.AnonymousID === FeedbackState.user.id).length
        };
        
        FeedbackState.stats = stats;
        
        const statsMap = {
            'total-feedbacks': stats.total,
            'pending-feedbacks': stats.pending,
            'resolved-feedbacks': stats.resolved,
            'my-feedbacks': stats.myFeedbacks
        };
        
        Object.entries(statsMap).forEach(([id, value]) => {
            const el = document.getElementById(id);
            if (el) el.textContent = value;
        });
    },

    // 显示通知
    showNotification(message, type = 'info') {
        // 简单的 alert 通知
        alert(message);
    },

    // 打开详情模态框
    openDetailModal(feedback) {
        const modal = document.getElementById('feedbackDetailModal');
        const contentDiv = document.getElementById('detailModalContent');
        if (!modal || !contentDiv) return;
        
        // 格式化日期
        const formatDate = (dateStr) => {
            if (!dateStr) return '未提供';
            try {
                const date = new Date(dateStr);
                return date.toLocaleString('zh-CN');
            } catch {
                return dateStr;
            }
        };
        
        const isTeacher = FeedbackState.user.role === 'teacher' || localStorage.getItem('user_role') === 'teacher';
        const isMyFeedback = feedback.AnonymousID === FeedbackState.user.id || feedback.AnonymousID === localStorage.getItem('user_id');
        
        // 构建模态框内容
        contentDiv.innerHTML = `
            <div style="margin-bottom: 20px;">
                <h3 style="font-size: 22px; margin: 0 0 15px 0; color: #333; word-break: break-word;">${feedback.Title || '无标题'}</h3>
                <div style="display: flex; gap: 12px; flex-wrap: wrap; margin-bottom: 20px;">
                    <span style="background: #ecf5ff; color: #409eff; padding: 4px 12px; border-radius: 20px; font-size: 13px;">${this.getTypeLabel(feedback.Type)}</span>
                    <span style="background: ${this.getStatusColor(feedback.Status)}20; color: ${this.getStatusColor(feedback.Status)}; padding: 4px 12px; border-radius: 20px; font-size: 13px;">${this.getStatusLabel(feedback.Status)}</span>
                    ${isMyFeedback ? '<span style="background: #fdf6ec; color: #e6a23c; padding: 4px 12px; border-radius: 20px; font-size: 13px;">我的反馈</span>' : ''}
                </div>
                <div style="background: #f8f9fa; border-radius: 12px; padding: 20px; margin-bottom: 25px; border: 1px solid #f0f0f0;">
                    <p style="margin: 0 0 15px 0; font-size: 15px; line-height: 1.7; color: #555; white-space: pre-wrap; word-break: break-word;">${feedback.Content || '无内容'}</p>
                    <div style="display: flex; justify-content: space-between; align-items: center; color: #999; font-size: 13px; border-top: 1px solid #eee; padding-top: 15px;">
                        <span>${feedback.AnonymousID || '匿名用户'}</span>
                        <span>${formatDate(feedback.CreatedAt)}</span>
                    </div>
                </div>
                
                <!-- 教师回复展示区域 -->
                ${feedback.TeacherResponse ? `
                <div style="background: #fef7e7; border-radius: 12px; padding: 20px; margin-bottom: 25px; border-left: 4px solid #e6a23c;">
                    <div style="display: flex; align-items: center; margin-bottom: 12px;">
                        <span style="background: #e6a23c; color: white; padding: 2px 10px; border-radius: 16px; font-size: 12px; font-weight: bold;">教师回复</span>
                        <span style="margin-left: 12px; color: #999; font-size: 12px;">${formatDate(feedback.RespondedAt)}</span>
                    </div>
                    <p style="margin: 0; color: #666; font-size: 14px; line-height: 1.6; white-space: pre-wrap; word-break: break-word;">${feedback.TeacherResponse}</p>
                </div>
                ` : ''}

                <!-- 教师回复输入区域（仅教师可见） -->
                ${isTeacher ? `
                <div style="margin-bottom: 25px; padding: 20px; background: #f0f9eb; border-radius: 12px; border: 1px dashed #67c23a;">
                    <label style="display: block; margin-bottom: 8px; color: #555; font-size: 14px; font-weight: 600;">📝 ${feedback.TeacherResponse ? '修改回复' : '发布回复'}</label>
                    <textarea id="teacherResponseInput" placeholder="输入您的回复内容..." style="width: 100%; padding: 12px; border: 1px solid #dcdfe6; border-radius: 8px; font-size: 14px; height: 100px; resize: vertical; margin-bottom: 10px;">${feedback.TeacherResponse || ''}</textarea>
                    <button id="submitResponseBtn" class="btn" style="padding: 8px 20px; background: #67c23a; border: none; color: white; font-weight: 500;">${feedback.TeacherResponse ? '更新回复' : '提交回复'}</button>
                </div>
                ` : ''}
                
                <!-- 操作按钮区域 -->
                <div style="border-top: 1px solid #eee; padding-top: 25px; margin-top: 10px;">
                    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                        <button id="modalLikeBtn" class="feedback-like-btn ${feedback.Liked ? 'liked' : ''}">
                            👍 点赞 (${feedback.LikeCount || 0})
                        </button>
                        <button id="modalCloseBtn" class="btn btn-secondary" style="padding: 8px 20px;">关闭</button>
                    </div>

                    ${isTeacher ? `
                    <div style="background: #f8fafc; border-radius: 12px; padding: 20px; border: 1px solid #e2e8f0;">
                        <h4 style="margin: 0 0 15px 0; font-size: 14px; color: #64748b; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px;">🛠️ 管理操作</h4>
                        <div style="display: flex; flex-wrap: wrap; gap: 10px; margin-bottom: 15px;">
                            <button id="modalStatusPendingBtn" class="feedback-status-btn ${feedback.Status === 'pending' ? 'active' : ''}" ${feedback.Status === 'pending' ? 'disabled' : ''}>⏳ 待处理</button>
                            <button id="modalStatusProcessingBtn" class="feedback-status-btn ${feedback.Status === 'processing' ? 'active' : ''}" ${feedback.Status === 'processing' ? 'disabled' : ''}>🔄 处理中</button>
                            <button id="modalStatusResolvedBtn" class="feedback-status-btn ${feedback.Status === 'resolved' ? 'active' : ''}" ${feedback.Status === 'resolved' ? 'disabled' : ''}>✅ 已解决</button>
                            <button id="modalStatusClosedBtn" class="feedback-status-btn ${feedback.Status === 'closed' ? 'active' : ''}" ${feedback.Status === 'closed' ? 'disabled' : ''}>🔒 已关闭</button>
                        </div>
                        <div style="border-top: 1px solid #e2e8f0; padding-top: 15px; display: flex; justify-content: flex-end;">
                            <button id="modalDeleteBtn" class="btn btn-danger" style="padding: 8px 16px; font-size: 13px;">🗑️ 删除反馈</button>
                        </div>
                    </div>
                    ` : ''}
                    
                    ${!isTeacher && isMyFeedback ? `
                    <div style="display: flex; justify-content: flex-end; margin-top: 15px;">
                        <button id="modalDeleteBtn" class="btn btn-danger" style="padding: 8px 16px; font-size: 13px;">🗑️ 删除我的反馈</button>
                    </div>
                    ` : ''}
                </div>
            </div>
        `;
        
        // 显示模态框
        modal.style.display = 'flex';
        
        // 绑定模态框内的按钮事件
        const closeBtn = document.getElementById('closeDetailModalBtn');
        const closeModalBtn = document.getElementById('modalCloseBtn');
        const likeBtn = document.getElementById('modalLikeBtn');
        
        // 关闭事件
        const closeModal = () => { modal.style.display = 'none'; };
        if (closeBtn) closeBtn.onclick = closeModal;
        if (closeModalBtn) closeModalBtn.onclick = closeModal;
        
        // 点击模态框背景关闭
        modal.onclick = (e) => {
            if (e.target === modal) closeModal();
        };
        
        // 点赞事件
        if (likeBtn) {
            likeBtn.onclick = async () => {
                try {
                    await FeedbackController.toggleLike(feedback.ID);
                    // 更新本地点赞数并刷新模态框显示
                    const updatedFeedback = await FeedbackAPI.fetchFeedbackDetail(feedback.ID);
                    this.openDetailModal(updatedFeedback);
                } catch (error) {
                    FeedbackUI.showNotification('操作失败', 'error');
                }
            };
        }
        
        // 教师操作：回复提交
        const submitResponseBtn = document.getElementById('submitResponseBtn');
        if (submitResponseBtn) {
            submitResponseBtn.onclick = async () => {
                const responseText = document.getElementById('teacherResponseInput').value.trim();
                if (!responseText) {
                    FeedbackUI.showNotification('请输入回复内容', 'warning');
                    return;
                }
                try {
                    await FeedbackController.respondFeedback(feedback.ID, responseText);
                    FeedbackUI.showNotification('回复成功', 'success');
                    // 重新加载反馈详情
                    const updatedFeedback = await FeedbackAPI.fetchFeedbackDetail(feedback.ID);
                    this.openDetailModal(updatedFeedback);
                    // 刷新列表
                    FeedbackController.loadFeedbacks();
                } catch (error) {
                    FeedbackUI.showNotification('回复失败', 'error');
                }
            };
        }
        
        // 教师操作：状态变更
        const statusMapping = {
            'modalStatusPendingBtn': 'pending',
            'modalStatusProcessingBtn': 'processing',
            'modalStatusResolvedBtn': 'resolved',
            'modalStatusClosedBtn': 'closed'
        };
        Object.entries(statusMapping).forEach(([btnId, status]) => {
            const btn = document.getElementById(btnId);
            if (btn) {
                btn.onclick = async () => {
                    try {
                        await FeedbackController.updateStatus(feedback.ID, status);
                        FeedbackUI.showNotification(`状态已更新为 ${this.getStatusLabel(status)}`, 'success');
                        // 重新加载反馈详情
                        const updatedFeedback = await FeedbackAPI.fetchFeedbackDetail(feedback.ID);
                        this.openDetailModal(updatedFeedback);
                        // 刷新列表
                        FeedbackController.loadFeedbacks();
                    } catch (error) {
                        FeedbackUI.showNotification('状态更新失败', 'error');
                    }
                };
            }
        });
        
        // 删除操作
        const deleteBtn = document.getElementById('modalDeleteBtn');
        if (deleteBtn) {
            deleteBtn.onclick = async () => {
                if (!confirm('确定要删除这条反馈吗？此操作不可撤销。')) return;
                try {
                    await FeedbackController.deleteFeedback(feedback.ID);
                    FeedbackUI.showNotification('删除成功', 'success');
                    closeModal();
                    // 刷新列表
                    FeedbackController.loadFeedbacks();
                } catch (error) {
                    FeedbackUI.showNotification('删除失败', 'error');
                }
            };
        }
    },

    // 关闭详情模态框（外部也可调用）
    closeDetailModal() {
        const modal = document.getElementById('feedbackDetailModal');
        if (modal) modal.style.display = 'none';
    },

    // 获取状态对应的颜色
    getStatusColor(status) {
        const map = {
            'pending': '#e6a23c',
            'processing': '#409eff',
            'resolved': '#67c23a',
            'closed': '#909399',
            'open': '#e6a23c'
        };
        return map[status] || '#909399';
    },

    // 辅助方法
    getTypeLabel(type) {
        const map = { 
            'bug': '🐛 Bug报告', 
            'feature': '✨ 功能建议', 
            'praise': '👍 点赞表扬', 
            'suggestion': '💡 学习建议', 
            'question': '❓ 问题咨询', 
            'other': '📝 其他' 
        };
        return map[type] || type || '📝 其他';
    },

    getStatusLabel(status) {
        const map = { 
            'pending': '⏳ 待处理', 
            'processing': '🔄 处理中', 
            'resolved': '✅ 已解决', 
            'closed': '🔒 已关闭' 
        };
        return map[status] || status || '⏳ 待处理';
    }
};

// ==================== 控制器 ====================
const FeedbackController = {
    // 初始化
    async init() {
        try {
            // 先绑定事件，确保即使加载数据失败，过滤器和搜索也能用
            this.bindEvents();

            FeedbackState.ui.isLoading = true;
            await this.loadFeedbacks();
            
            // 检查 URL 参数，如果有 id，自动打开详情
            const urlParams = new URLSearchParams(window.location.search);
            const feedbackId = urlParams.get('id');
            if (feedbackId) {
                this.viewFeedbackDetail(parseInt(feedbackId, 10));
            }
            
            FeedbackState.ui.isLoading = false;
        } catch (error) {
            console.error('FeedbackController init failed:', error);
            FeedbackState.ui.isLoading = false;
            // 即使加载失败也更新一次统计（显示0）
            FeedbackUI.updateStats([]);
            FeedbackUI.showNotification('初始化失败，请刷新页面重试', 'error');
        }
    },

    // 加载反馈列表
    async loadFeedbacks(params = null) {
        try {
            // 如果传了参数，更新全局状态以保持同步
            if (params) {
                if (params.type !== undefined) {
                    FeedbackState.filter.type = params.type;
                    const el = document.getElementById('filter-type');
                    if (el) el.value = params.type;
                }
                if (params.status !== undefined) {
                    FeedbackState.filter.status = params.status;
                    const el = document.getElementById('filter-status');
                    if (el) el.value = params.status;
                }
                if (params.search !== undefined) {
                    FeedbackState.filter.search = params.search;
                    const el = document.getElementById('search-feedback');
                    if (el) el.value = params.search;
                }
            }

            // 构建最终抓取参数
            const fetchParams = {
                type: FeedbackState.filter.type,
                status: FeedbackState.filter.status,
                search: FeedbackState.filter.search
            };
            
            const feedbacks = await FeedbackAPI.fetchFeedbacks(fetchParams);
            FeedbackState.feedbacks = feedbacks;
            // 过滤 / 搜索时重置到第一页
            FeedbackState.ui.currentPage = 1;
            FeedbackUI.renderFeedbacks(feedbacks);
            FeedbackUI.updateStats(feedbacks);
        } catch (error) {
            console.error('加载反馈失败:', error);
        }
    },

    // 显示发布反馈表单
    showFeedbackForm() {
        const form = document.getElementById('feedbackForm');
        if (form) {
            form.style.display = 'block';
            FeedbackState.ui.showForm = true;
        }
    },

    // 隐藏发布反馈表单
    hideFeedbackForm() {
        const form = document.getElementById('feedbackForm');
        if (form) {
            form.style.display = 'none';
            FeedbackState.ui.showForm = false;
        }
    },

    // 提交新反馈
    async submitFeedback() {
        const type = document.getElementById('fbType')?.value;
        const title = document.getElementById('fbTitle')?.value;
        const content = document.getElementById('fbContent')?.value;
        
        if (!title || !content) {
            alert('请填写标题和内容');
            return;
        }
        
        try {
            await FeedbackAPI.createFeedback({
                type,
                title,
                content,
                anonymous_id: FeedbackState.user.id // 统一使用当前用户ID
            });
            alert('反馈发布成功！');
            this.hideFeedbackForm();
            // 重置表单
            if (document.getElementById('fbTitle')) document.getElementById('fbTitle').value = '';
            if (document.getElementById('fbContent')) document.getElementById('fbContent').value = '';
            // 刷新列表
            await this.loadFeedbacks();
        } catch (error) {
            console.error('发布反馈失败:', error);
            alert('发布失败: ' + error.message);
        }
    },

    // 分页切换
    goToPage(page) {
        const total = FeedbackState.ui.totalPages || 1;
        if (page < 1 || page > total) return;
        FeedbackState.ui.currentPage = page;
        FeedbackUI.renderFeedbacks(FeedbackState.feedbacks);
    },

    // 切换点赞
    async toggleLike(id) {
        try {
            await FeedbackAPI.likeFeedback(id);
            // 刷新当前列表
            await this.loadFeedbacks();
            // 如果是在详情模态框中，重新获取详情以更新显示
            if (FeedbackState.ui.showDetailModal && FeedbackState.ui.currentFeedbackId === id) {
                const updated = await FeedbackAPI.fetchFeedbackDetail(id);
                FeedbackUI.openDetailModal(updated);
            }
            FeedbackUI.showNotification('点赞成功', 'success');
        } catch (error) {
            console.error('点赞失败:', error);
            FeedbackUI.showNotification('点赞失败', 'error');
        }
    },

    // 教师回复反馈
    async respondFeedback(id, responseText) {
        try {
            await FeedbackAPI.request(`/api/feedback/${id}/respond`, {
                method: 'POST',
                body: JSON.stringify({ response: responseText })
            });
            return true;
        } catch (error) {
            console.error('回复失败:', error);
            throw error;
        }
    },

    // 教师更新反馈状态
    async updateStatus(id, status) {
        try {
            await FeedbackAPI.request(`/api/feedback/${id}/status`, {
                method: 'PUT',
                body: JSON.stringify({ status })
            });
            return true;
        } catch (error) {
            console.error('更新状态失败:', error);
            throw error;
        }
    },

    // 教师删除反馈
    async deleteFeedback(id) {
        try {
            await FeedbackAPI.request(`/api/feedback/${id}`, {
                method: 'DELETE'
            });
            return true;
        } catch (error) {
            console.error('删除失败:', error);
            throw error;
        }
    },

    // 显示状态切换菜单（教师端使用）
    showStatusMenu(feedbackId, currentStatus, evt) {
        const button = evt && (evt.currentTarget || evt.target);
        if (!button) return;

        // 如果已有菜单，先移除
        let existing = document.getElementById('feedback-status-dropdown');
        if (existing) {
            existing.remove();
            if (existing.dataset.forId === String(feedbackId)) return;
        }

        const statuses = [
            { value: 'pending', text: '待处理' },
            { value: 'processing', text: '处理中' },
            { value: 'resolved', text: '已解决' },
            { value: 'closed', text: '已关闭' }
        ];

        const rect = button.getBoundingClientRect();
        const menu = document.createElement('div');
        menu.id = 'feedback-status-dropdown';
        menu.dataset.forId = String(feedbackId);
        Object.assign(menu.style, {
            position: 'fixed',
            top: `${rect.bottom + 4}px`,
            left: `${rect.left}px`,
            background: '#fff',
            border: '1px solid #e5e7eb',
            boxShadow: '0 8px 16px rgba(15,23,42,0.15)',
            borderRadius: '8px',
            zIndex: '2000',
            minWidth: '140px',
            padding: '4px 0'
        });

        menu.innerHTML = statuses.map(s => `
            <button type="button" data-status="${s.value}"
                    style="width: 100%; padding: 8px 16px; background: ${s.value === currentStatus ? '#eff6ff' : 'transparent'};
                           border: none; text-align: left; font-size: 13px; color: #374151; cursor: pointer;"
                    onmouseover="this.style.background='#eff6ff'"
                    onmouseout="this.style.background='${s.value === currentStatus ? '#eff6ff' : 'transparent'}'">
                ${s.text}
            </button>
        `).join('');

        document.body.appendChild(menu);

        const onOutsideClick = (e) => {
            if (!menu.contains(e.target) && !button.contains(e.target)) {
                menu.remove();
                document.removeEventListener('click', onOutsideClick, true);
            }
        };

        menu.addEventListener('click', async (e) => {
            const btn = e.target.closest('button[data-status]');
            if (!btn) return;
            const status = btn.getAttribute('data-status');
            menu.remove();
            document.removeEventListener('click', onOutsideClick, true);
            
            if (confirm(`确定要将状态更新为"${btn.textContent.trim()}"吗？`)) {
                try {
                    await this.updateStatus(feedbackId, status);
                    FeedbackUI.showNotification('状态更新成功', 'success');
                    await this.loadFeedbacks();
                } catch (error) {
                    FeedbackUI.showNotification('状态更新失败', 'error');
                }
            }
        });

        setTimeout(() => document.addEventListener('click', onOutsideClick, true), 0);
    },

    // 显示回复对话框（教师端使用）
    showRespondModal(feedbackId) {
        // 直接使用 viewFeedbackDetail 打开详情模态框，详情模态框中已有回复功能
        this.viewFeedbackDetail(feedbackId);
    },

    // 查看反馈详情（使用模态框）
    async viewFeedbackDetail(id) {
        try {
            // 先从本地状态中查找
            let feedback = FeedbackState.feedbacks.find(f => f.ID === id);
            
            // 如果本地没有或者缺少完整内容（如 Content, TeacherResponse），则调用API获取详情
            if (!feedback || !feedback.Content) {
                feedback = await FeedbackAPI.fetchFeedbackDetail(id);
            }
            
            if (feedback) {
                FeedbackUI.openDetailModal(feedback);
            } else {
                FeedbackUI.showNotification('无法获取反馈详情', 'error');
            }
        } catch (error) {
            console.error('获取反馈详情失败:', error);
            FeedbackUI.showNotification('获取反馈详情失败', 'error');
        }
    },

    // 绑定事件
    bindEvents() {
        // 过滤事件 - 类型
        const typeFilter = document.getElementById('filter-type');
        if (typeFilter) {
            typeFilter.addEventListener('change', (e) => {
                this.loadFeedbacks({ 
                    type: e.target.value,
                    status: document.getElementById('filter-status')?.value || '',
                    search: document.getElementById('search-feedback')?.value || ''
                });
            });
        }

        // 过滤事件 - 状态
        const statusFilter = document.getElementById('filter-status');
        if (statusFilter) {
            statusFilter.addEventListener('change', (e) => {
                this.loadFeedbacks({ 
                    type: document.getElementById('filter-type')?.value || '',
                    status: e.target.value,
                    search: document.getElementById('search-feedback')?.value || ''
                });
            });
        }

        // 搜索事件
        const searchInput = document.getElementById('search-feedback');
        const searchBtn = document.getElementById('search-btn');
        
        if (searchBtn && searchInput) {
            searchBtn.onclick = () => {
                this.loadFeedbacks({ 
                    type: document.getElementById('filter-type')?.value || '',
                    status: document.getElementById('filter-status')?.value || '',
                    search: searchInput.value
                });
            };
            
            // 回车搜索
            searchInput.onkeypress = (e) => {
                if (e.key === 'Enter') {
                    this.loadFeedbacks({ 
                        type: document.getElementById('filter-type')?.value || '',
                        status: document.getElementById('filter-status')?.value || '',
                        search: searchInput.value
                    });
                }
            };
        }

        // 使用事件委托监听反馈卡片上的详情链接点击（防止innerHTML覆盖后事件失效）
        const feedbackList = document.getElementById('feedback-list');
        if (feedbackList) {
            feedbackList.addEventListener('click', (e) => {
                // 查找被点击的 <a> 标签，其 onclick 属性包含 viewFeedbackDetail
                const targetLink = e.target.closest('a[href="javascript:void(0)"]');
                if (targetLink && targetLink.hasAttribute('onclick')) {
                    const onclickAttr = targetLink.getAttribute('onclick');
                    const match = onclickAttr.match(/FeedbackController\.viewFeedbackDetail\((\d+)\)/);
                    if (match && match[1]) {
                        e.preventDefault();
                        const id = parseInt(match[1], 10);
                        FeedbackController.viewFeedbackDetail(id);
                    }
                }
            });
        }
    }
};

// ==================== 初始化 ====================
(function initFeedbackModule() {
    // 等待 DOM 加载完成
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            FeedbackController.init();
        });
    } else {
        FeedbackController.init();
    }
})();

// 导出全局接口
window.FeedbackController = FeedbackController;
window.FeedbackState = FeedbackState;