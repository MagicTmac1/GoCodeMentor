window.App = {
    user: {
        id: null,
        name: null,
        role: null,
    },
    init(userData) {
        this.user.id = userData.id;
        this.user.name = userData.name;
        this.user.role = userData.role;

        if (window.FeedbackState) {
            window.FeedbackState.user = this.user;
        }

        document.getElementById('userName').textContent = this.user.name || (this.user.role === 'teacher' ? '教师' : '学生');

        if (this.user.role === 'admin') {
            const adminTab = document.getElementById('adminTab');
            if(adminTab) adminTab.style.display = 'block';
        }
    },
    logout() {
        sessionStorage.clear();
        localStorage.clear();
        document.cookie = "user_id=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
        document.cookie = "user_role=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
        document.cookie = "user_name=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
        window.location.replace('/login');
    },
    _tabToSectionId(tab) {
        const parts = tab.split('-');
        const capitalized = parts.map((part, index) => {
            if (index === 0) return part;
            return part.charAt(0).toUpperCase() + part.slice(1);
        });
        return capitalized.join('') + 'Section';
    },
    switchTab(tab) {
        document.querySelectorAll('.nav-tab').forEach(t => t.classList.remove('active'));
        const tabElement = Array.from(document.querySelectorAll('.nav-tab')).find(t => t.getAttribute('onclick').includes(`'${tab}'`));
        if (tabElement) tabElement.classList.add('active');

        const sections = ['courseDetailsSection', 'wisdomGraphSection', 'classesSection', 'assignmentsSection', 'feedbackSection', 'resourcesSection'];
        const activeSectionId = this._tabToSectionId(tab);

        sections.forEach(sectionId => {
            const sectionElement = document.getElementById(sectionId);
            if (sectionElement) {
                sectionElement.style.display = sectionId === activeSectionId ? 'block' : 'none';
            }
        });

        // Lazy load content
        if (tab === 'classes' && !window.classesLoaded) {
            if(typeof loadClasses === 'function') loadClasses();
            window.classesLoaded = true;
        } else if (tab === 'assignments' && !window.assignmentsLoaded) {
            if(typeof loadAssignments === 'function') loadAssignments();
            window.assignmentsLoaded = true;
        } else if (tab === 'wisdom-graph' && !window.wisdomGraphLoaded) {
            if(typeof loadWisdomGraph === 'function') loadWisdomGraph();
            window.wisdomGraphLoaded = true;
        } else if (tab === 'feedback' && !window.feedbackLoaded) {
            window.FeedbackController?.loadFeedbacks();
            window.feedbackLoaded = true;
        } else if (tab === 'resources' && !window.resourcesInitialLoad) {
            if(typeof loadResources === 'function') loadResources();
            window.resourcesInitialLoad = true;
        }
    }
};

async function loadWisdomGraph() {
    const chartDom = document.getElementById('wisdomGraphContainer');
    if (!chartDom) return;
    const myChart = echarts.init(chartDom);
    myChart.showLoading();

    try {
        const response = await fetch('/api/wisdom-graph');
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        const graphData = await response.json();

        const option = {
            title: {
                text: 'Go语言知识图谱',
                subtext: '知识点关联',
                top: 'top',
                left: 'center'
            },
            tooltip: {},
            legend: [{
                data: graphData.categories.map(a => a.name),
                orient: 'vertical',
                left: 'left'
            }],
            animationDuration: 1500,
            animationEasingUpdate: 'quinticInOut',
            series: [
                {
                    name: 'Go Wisdom Graph',
                    type: 'graph',
                    layout: 'force',
                    data: graphData.nodes,
                    links: graphData.links,
                    categories: graphData.categories,
                    roam: true,
                    label: {
                        show: true,
                        position: 'right',
                        formatter: '{b}'
                    },
                    lineStyle: {
                        color: 'source',
                        curveness: 0.3
                    },
                    emphasis: {
                        focus: 'adjacency',
                        lineStyle: {
                            width: 10
                        }
                    },
                    force: {
                        repulsion: 200,
                        edgeLength: 80
                    }
                }
            ]
        };
        myChart.hideLoading();
        myChart.setOption(option);
    } catch (error) {
        myChart.hideLoading();
        chartDom.innerHTML = '<div style="text-align: center; padding-top: 50px; color: red;">知识图谱加载失败，请稍后重试。</div>';
        console.error('Failed to load wisdom graph:', error);
    }
}

// --- Resource Recommendations Logic ---
let resourcesInitialLoad = false;

function showAddResourceModal() {
    const modal = document.getElementById('addResourceModal');
    if (modal) modal.style.display = 'flex';
}

function hideAddResourceModal() {
    const modal = document.getElementById('addResourceModal');
    if(modal) modal.style.display = 'none';
    document.getElementById('newResourceTitle').value = '';
    document.getElementById('newResourceUrl').value = '';
    document.getElementById('newResourceDesc').value = '';
}

async function submitNewResource() {
    const title = document.getElementById('newResourceTitle').value;
    const url = document.getElementById('newResourceUrl').value;
    const description = document.getElementById('newResourceDesc').value;
    const category = document.getElementById('newResourceCategory').value;

    if (!title || !url || !category) {
        alert('标题、链接和分类是必填项！');
        return;
    }

    const resourceId = title.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '');

    try {
        const res = await fetch('/api/resources', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-User-ID': window.App.user.id,
                'X-User-Role': window.App.user.role
            },
            body: JSON.stringify({ 
                resourceId, 
                title, 
                url, 
                description, 
                category, 
                iconURL: `https://www.google.com/s2/favicons?sz=64&domain=${new URL(url).hostname}`
            })
        });

        if (res.ok) {
            alert('新资源添加成功！');
            hideAddResourceModal();
            loadResources();
        } else {
            const data = await res.json();
            alert('添加失败: ' + (data.error || '未知错误'));
        }
    } catch (e) {
        alert('网络请求失败: ' + e.message);
        console.error('Error submitting new resource:', e);
    }
}

async function loadResources() {
    try {
        const res = await fetch('/api/resources', {
             headers: { 'X-User-ID': window.App.user.id, 'X-User-Role': window.App.user.role }
        });
        if (!res.ok) throw new Error(`获取资源失败: ${res.status}`);
        
        const resources = await res.json();

        document.querySelectorAll('.resource-list').forEach(list => list.innerHTML = '');

        if (!resources || resources.length === 0) {
             document.querySelector('#cat-official .resource-list').innerHTML = '<p>暂无资源，快去添加吧！</p>';
             document.getElementById('leaderboard-list').innerHTML = '<p style="color: #999; font-size: 13px;">暂无推荐，快去点赞吧！</p>';
             return;
        }

        resources.forEach(resource => {
            const listContainer = document.querySelector(`#cat-${resource.category} .resource-list`);
            if (listContainer) {
                const deleteButtonHTML = window.App.user.role === 'teacher' || window.App.user.role === 'admin' ? 
                    `<button class="btn btn-secondary delete-btn" style="padding: 8px 12px; font-size: 12px; background: #f8d7da; color: #721c24; border-color: #f5c6cb;" onclick="event.stopPropagation(); deleteResource('${resource.resourceId}')">🗑️</button>` : '';

                const resourceHTML = `
                    <li class="resource-item" data-resource-id="${resource.resourceId}">
                        <img src="${resource.iconURL || 'https://www.google.com/s2/favicons?sz=64&domain=go.dev'}" alt="${resource.title} icon">
                        <div class="text-content">
                            <a href="${resource.url}" target="_blank">${resource.title}</a>
                            <p>${resource.description || '暂无描述'}</p>
                        </div>
                        <div class="like-container" onclick="toggleLike(this, '${resource.resourceId}')">
                            <button class="like-btn">👍</button>
                            <span class="like-count">0</span>
                        </div>
                        ${deleteButtonHTML}
                    </li>
                `;
                listContainer.insertAdjacentHTML('beforeend', resourceHTML);
            }
        });

        loadResourceStats();

    } catch(e) {
        console.error("Failed to load resources:", e);
        document.querySelector('#cat-official .resource-list').innerHTML = `<p style="color:red;">加载资源失败: ${e.message}</p>`;
    }
}

async function loadResourceStats() {
    const leaderboardList = document.getElementById('leaderboard-list');
    try {
        const res = await fetch('/api/resources/stats', {
            headers: { 'X-User-ID': window.App.user.id, 'X-User-Role': window.App.user.role }
        });
        if (!res.ok) throw new Error(`请求失败，状态码: ${res.status}`);

        const stats = await res.json();

        document.querySelectorAll('.resource-item').forEach(item => {
            const resourceId = item.dataset.resourceId;
            if (!resourceId) return;

            const likeCount = stats.likeCounts[resourceId] || 0;
            const userLiked = stats.userLikes[resourceId] || false;

            const countElement = item.querySelector('.like-count');
            const containerElement = item.querySelector('.like-container');

            if (countElement) countElement.textContent = likeCount;
            if (containerElement) {
                containerElement.classList.toggle('liked', userLiked);
            }
        });

        if (stats.leaderboard && stats.leaderboard.length > 0) {
            leaderboardList.innerHTML = stats.leaderboard.map((item, index) => {
                const resourceElement = document.querySelector(`.resource-item[data-resource-id='${item.resourceId}']`);
                if (!resourceElement) return '';

                const title = resourceElement.querySelector('.text-content a').textContent;
                const href = resourceElement.querySelector('.text-content a').href;
                const iconSrc = resourceElement.querySelector('img').src;
                const rankClass = index < 3 ? `top-${index + 1}` : '';

                return `
                    <div class="leaderboard-item">
                        <span class="leaderboard-rank ${rankClass}">${index + 1}</span>
                        <img src="${iconSrc}" alt="Icon">
                        <div class="leaderboard-item-info">
                            <a href="${href}" target="_blank" title="${title}">${title}</a>
                            <div class="like-count-lb">👍 ${item.likeCount}</div>
                        </div>
                    </div>
                `;
            }).join('');
        } else {
            leaderboardList.innerHTML = '<p style="color: #999; font-size: 13px;">暂无推荐，快去点赞吧！</p>';
        }

    } catch (error) {
        console.error('Error loading resource stats:', error);
        leaderboardList.innerHTML = `<p style="color: #dc3545; font-size: 13px; padding: 10px;">排行榜加载失败: ${error.message}</p>`;
    }
}

async function toggleLike(element, resourceId) {
    try {
        const res = await fetch(`/api/resources/${resourceId}/like`, {
            method: 'POST',
            headers: { 'X-User-ID': window.App.user.id, 'X-User-Role': window.App.user.role }
        });

        if (!res.ok) throw new Error('Failed to toggle like');

        const result = await res.json();

        const countElement = element.querySelector('.like-count');
        countElement.textContent = result.newCount;
        element.classList.toggle('liked', result.liked);
        
        setTimeout(loadResourceStats, 500);

    } catch (error) {
        console.error('Error toggling like:', error);
        alert('操作失败，请稍后重试。');
    }
}

function switchResourceCategory(event, category) {
    event.preventDefault();
    document.querySelectorAll('.resources-sidebar-item').forEach(item => item.classList.remove('active'));
    event.target.classList.add('active');
    document.querySelectorAll('.resource-category-content').forEach(content => content.classList.remove('active'));
    document.getElementById('cat-' + category).classList.add('active');
}

async function deleteResource(resourceId) {
    if (!confirm(`确定要删除这个资源吗？此操作不可撤销。`)) {
        return;
    }

    try {
        const res = await fetch(`/api/resources/${resourceId}`, {
            method: 'DELETE',
            headers: {
                'X-User-ID': window.App.user.id,
                'X-User-Role': window.App.user.role
            }
        });

        if (res.ok) {
            alert('资源删除成功！');
            loadResources();
        } else {
            const data = await res.json();
            alert('删除失败: ' + (data.error || '未知错误'));
        }
    } catch (e) {
        alert('网络请求失败: ' + e.message);
        console.error('Error deleting resource:', e);
    }
}
