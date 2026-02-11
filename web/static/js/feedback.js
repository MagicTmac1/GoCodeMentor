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
            <div class="feedback-card feedback-row ${feedback.Status || 'pending'}" data-id="${feedback.ID}">
                <div class="feedback-row-main">
                    <div class="feedback-row-title">
                        <a href="javascript:void(0)" onclick="FeedbackController.viewFeedbackDetail(${feedback.ID})" class="feedback-title-link">
                            ${feedback.Title || '无标题'}
                        </a>
                        <span class="feedback-label feedback-label--${feedback.Type || 'other'}">${typeLabel}</span>
                        <span class="feedback-status-pill feedback-status-pill--${feedback.Status || 'pending'}">${statusLabel}</span>
                        ${isMyFeedback ? '<span class="feedback-mine-pill">我的反馈</span>' : ''}
                    </div>
                    ${snippet ? `<div class="feedback-row-snippet">${snippet}</div>` : ''}
                    <div class="feedback-row-meta">
                        <span>${feedback.AnonymousID || '匿名用户'}</span>
                        <span>${dateStr}</span>
                        <span>👍 ${feedback.LikeCount || 0}</span>
                    </div>
                </div>
                <div class="feedback-row-actions">
                    <button class="feedback-like-chip ${likeClass}" onclick="FeedbackController.toggleLike(${feedback.ID})" title="点赞">
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
        
        const isTeacher = FeedbackState.user.role === 'teacher';
        const isMyFeedback = feedback.AnonymousID === FeedbackState.user.id;
        
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
                
                <!-- 教师回复区域 -->
                ${feedback.TeacherResponse ? `
                <div style="background: #fef7e7; border-radius: 12px; padding: 20px; margin-bottom: 25px; border-left: 4px solid #e6a23c;">
                    <div style="display: flex; align-items: center; margin-bottom: 12px;">
                        <span style="background: #e6a23c; color: white; padding: 2px 10px; border-radius: 16px; font-size: 12px; font-weight: bold;">教师回复</span>
                        <span style="margin-left: 12px; color: #999; font-size: 12px;">${formatDate(feedback.RespondedAt)}</span>
                    </div>
                    <p style="margin: 0; color: #666; font-size: 14px; line-height: 1.6; white-space: pre-wrap; word-break: break-word;">${feedback.TeacherResponse}</p>
                </div>
                ` : (isTeacher ? `
                <div style="margin-bottom: 25px;">
                    <label style="display: block; margin-bottom: 8px; color: #555; font-size: 14px; font-weight: 500;">📝 教师回复</label>
                    <textarea id="teacherResponseInput" placeholder="输入回复内容..." style="width: 100%; padding: 12px; border: 1px solid #e0e0e0; border-radius: 8px; font-size: 14px; height: 100px; resize: vertical;">${feedback.TeacherResponse || ''}</textarea>
                    <button id="submitResponseBtn" class="btn" style="margin-top: 12px; padding: 8px 20px; background: #e6a23c; border: none;">发布回复</button>
                </div>
                ` : '')}
                
                <!-- 操作按钮区域 -->
                <div style="display: flex; gap: 12px; flex-wrap: wrap; border-top: 1px solid #eee; padding-top: 25px; margin-top: 10px;">
                    <!-- 通用：点赞按钮 -->
                    <button id="modalLikeBtn" class="feedback-like-btn ${feedback.Liked ? 'liked' : ''}" style="margin-right: auto;">
                        <i class="icon-heart"></i> 点赞 (${feedback.LikeCount || 0})
                    </button>
                    
                    <!-- 教师特有按钮 -->
                    ${isTeacher ? `
                        <button id="modalStatusPendingBtn" class="feedback-status-btn" ${feedback.Status === 'pending' ? 'disabled' : ''}>⏳ 标记待处理</button>
                        <button id="modalStatusProcessingBtn" class="feedback-status-btn" ${feedback.Status === 'processing' ? 'disabled' : ''}>🔄 标记处理中</button>
                        <button id="modalStatusResolvedBtn" class="feedback-status-btn" ${feedback.Status === 'resolved' ? 'disabled' : ''}>✅ 标记已解决</button>
                        <button id="modalStatusClosedBtn" class="feedback-status-btn" ${feedback.Status === 'closed' ? 'disabled' : ''}>🔒 标记已关闭</button>
                        <button id="modalDeleteBtn" class="btn btn-danger">🗑️ 删除反馈</button>
                    ` : ''}
                    
                    <!-- 非教师（学生/匿名）且是自己的反馈：可删除（或取消点赞已在通用按钮） -->
                    ${!isTeacher && isMyFeedback ? `
                        <button id="modalDeleteBtn" class="btn btn-danger">🗑️ 删除我的反馈</button>
                    ` : ''}
                    
                    <button id="modalCloseBtn" class="btn btn-secondary" style="margin-left: auto;">关闭</button>
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
                if (!confirm('确定要删除这条反馈吗？')) return;
                try {
                    await FeedbackController.deleteFeedback(feedback.ID);
                    FeedbackUI.showNotification('删除成功', 'success');
                    closeModal();
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
        const map = { 'bug': '故障', 'suggestion': '建议', 'question': '疑问', 'other': '其他' };
        return map[type] || type || '其他';
    },

    getStatusLabel(status) {
        const map = { 'open': '待处理', 'pending': '待处理', 'processing': '处理中', 'resolved': '已解决', 'closed': '已关闭' };
        return map[status] || status || '待处理';
    }
};

// ==================== 控制器 ====================
const FeedbackController = {
    // 初始化
    async init() {
        try {
            FeedbackState.ui.isLoading = true;
            await this.loadFeedbacks();
            this.bindEvents();
            // 检查 URL 参数，如果有 id，自动打开详情
            const urlParams = new URLSearchParams(window.location.search);
            const feedbackId = urlParams.get('id');
            if (feedbackId) {
                this.viewFeedbackDetail(parseInt(feedbackId, 10));
            }
            FeedbackState.ui.isLoading = false;
        } catch (error) {
            console.error('反馈模块初始化失败:', error);
            FeedbackUI.showNotification('初始化失败，请刷新页面重试', 'error');
        }
    },

    // 加载反馈列表
    async loadFeedbacks(params = {}) {
        try {
            const feedbacks = await FeedbackAPI.fetchFeedbacks(params);
            FeedbackState.feedbacks = feedbacks;
            // 过滤 / 搜索时重置到第一页
            FeedbackState.ui.currentPage = 1;
            FeedbackUI.renderFeedbacks(feedbacks);
            FeedbackUI.updateStats(feedbacks);
        } catch (error) {
            console.error('加载反馈失败:', error);
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
            await this.loadFeedbacks(); // 重新加载以更新点赞状态
            FeedbackUI.showNotification('点赞成功', 'success');
        } catch (error) {
            console.error('点赞失败:', error);
            FeedbackUI.showNotification('点赞失败', 'error');
        }
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
                FeedbackState.filter.type = e.target.value;
                this.loadFeedbacks({ type: e.target.value });
            });
        }

        // 过滤事件 - 状态
        const statusFilter = document.getElementById('filter-status');
        if (statusFilter) {
            statusFilter.addEventListener('change', (e) => {
                FeedbackState.filter.status = e.target.value;
                this.loadFeedbacks({ 
                    type: FeedbackState.filter.type,
                    status: e.target.value,
                    search: FeedbackState.filter.search
                });
            });
        }

        // 搜索事件
        const searchInput = document.getElementById('search-feedback');
        const searchBtn = document.getElementById('search-btn');
        
        if (searchBtn && searchInput) {
            searchBtn.addEventListener('click', () => {
                FeedbackState.filter.search = searchInput.value;
                this.loadFeedbacks({ 
                    type: FeedbackState.filter.type,
                    status: FeedbackState.filter.status,
                    search: searchInput.value
                });
            });
            
            // 回车搜索
            searchInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    FeedbackState.filter.search = searchInput.value;
                    this.loadFeedbacks({ 
                        type: FeedbackState.filter.type,
                        status: FeedbackState.filter.status,
                        search: searchInput.value
                    });
                }
            });
        }

        // 新建反馈按钮
        const newBtn = document.getElementById('new-feedback-btn');
        if (newBtn) {
            newBtn.addEventListener('click', () => {
                FeedbackState.ui.showForm = true;
                // 简化版：提示用户功能开发中
                alert('新建反馈功能将在后续版本支持');
            });
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

// 导出缺少的控制器方法（需要在实际项目中实现，此处先给出空实现避免报错）
FeedbackController.respondFeedback = FeedbackController.respondFeedback || async function(id, response) {
    // 调用后端 /api/feedback/:id/respond 接口
    await FeedbackAPI.request(`/api/feedback/${id}/respond`, {
        method: 'POST',
        body: JSON.stringify({ response })
    });
};

FeedbackController.updateStatus = FeedbackController.updateStatus || async function(id, status) {
    // 调用后端 /api/feedback/:id/status 接口 (PUT)
    await FeedbackAPI.request(`/api/feedback/${id}/status`, {
        method: 'PUT',
        body: JSON.stringify({ status })
    });
};

FeedbackController.deleteFeedback = FeedbackController.deleteFeedback || async function(id) {
    // 调用后端 /api/feedback/:id 接口 (DELETE)
    await FeedbackAPI.request(`/api/feedback/${id}`, {
        method: 'DELETE'
    });
};

FeedbackController.toggleLike = FeedbackController.toggleLike || async function(id) {
    // 已实现，此处保留防止重复定义
};

// 导出全局接口
window.FeedbackController = FeedbackController;
window.FeedbackState = FeedbackState;