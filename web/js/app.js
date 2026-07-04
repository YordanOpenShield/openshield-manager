// OpenShield Manager Web Interface
// API Client and UI Controller

const API_BASE = '/api';

// State
let currentAgentId = null;
let currentAgentTools = [];
let selectedAgents = new Set();

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    initNavigation();
    initEventListeners();
    loadDashboardData();
});

// Navigation
function initNavigation() {
    const navLinks = document.querySelectorAll('.nav-links li');
    navLinks.forEach(link => {
        link.addEventListener('click', () => {
            const page = link.dataset.page;
            switchPage(page);
            
            navLinks.forEach(l => l.classList.remove('active'));
            link.classList.add('active');
        });
    });
}

function switchPage(page) {
    // Hide all pages
    document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
    
    // Show selected page
    const pageEl = document.getElementById(`${page}-page`);
    if (pageEl) {
        pageEl.classList.add('active');
    }
    
    // Update page title
    const titles = {
        dashboard: 'Dashboard',
        agents: 'Agents',
        jobs: 'Jobs',
        tasks: 'Tasks',
        queries: 'Queries',
        executions: 'Query Executions'
    };
    document.getElementById('page-title').textContent = titles[page] || page;
    
    // Load page data
    switch(page) {
        case 'dashboard':
            loadDashboardData();
            break;
        case 'agents':
            loadAgents();
            break;
        case 'jobs':
            loadJobs();
            break;
        case 'tasks':
            loadTasks();
            break;
        case 'queries':
            loadQueries();
            break;
        case 'executions':
            loadExecutions();
            break;
    }
}

// Event Listeners
function initEventListeners() {
    // Refresh button
    document.getElementById('refresh-btn').addEventListener('click', () => {
        const activePage = document.querySelector('.page.active').id.replace('-page', '');
        switchPage(activePage);
        showToast('Data refreshed', 'success');
    });
    
    // Forms
    document.getElementById('create-job-form').addEventListener('submit', handleCreateJob);
    document.getElementById('assign-task-form').addEventListener('submit', handleAssignTask);
    document.getElementById('create-query-form').addEventListener('submit', handleCreateQuery);
    document.getElementById('run-query-form').addEventListener('submit', handleRunQuery);
    document.getElementById('execute-tool-form').addEventListener('submit', handleExecuteTool);
    
    // Select all agents checkbox
    document.getElementById('select-all-agents').addEventListener('change', (e) => {
        const checkboxes = document.querySelectorAll('.agent-checkbox');
        checkboxes.forEach(cb => {
            cb.checked = e.target.checked;
            const agentId = cb.dataset.agentId;
            if (e.target.checked) {
                selectedAgents.add(agentId);
            } else {
                selectedAgents.delete(agentId);
            }
        });
    });
    
    // Tool name change - update actions
    document.getElementById('tool-name-select').addEventListener('change', updateToolActions);
}

// API Functions
async function apiGet(endpoint) {
    try {
        const response = await fetch(`${API_BASE}${endpoint}`);
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return await response.json();
    } catch (error) {
        console.error(`API GET ${endpoint} failed:`, error);
        showToast(`Failed to load data: ${error.message}`, 'error');
        throw error;
    }
}

async function apiPost(endpoint, data) {
    try {
        const response = await fetch(`${API_BASE}${endpoint}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || `HTTP ${response.status}`);
        }
        return await response.json();
    } catch (error) {
        console.error(`API POST ${endpoint} failed:`, error);
        showToast(`Failed: ${error.message}`, 'error');
        throw error;
    }
}

async function apiDelete(endpoint) {
    try {
        const response = await fetch(`${API_BASE}${endpoint}`, {
            method: 'DELETE'
        });
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return await response.json();
    } catch (error) {
        console.error(`API DELETE ${endpoint} failed:`, error);
        showToast(`Failed: ${error.message}`, 'error');
        throw error;
    }
}

// Dashboard
async function loadDashboardData() {
    try {
        const [agents, jobs, tasks] = await Promise.all([
            apiGet('/agents/list'),
            apiGet('/jobs/list'),
            apiGet('/tasks/list')
        ]);
        
        // Update stats
        document.getElementById('total-agents').textContent = agents.length;
        document.getElementById('connected-agents').textContent = 
            agents.filter(a => a.state === 'CONNECTED').length;
        document.getElementById('total-jobs').textContent = jobs.length;
        document.getElementById('pending-tasks').textContent = 
            tasks.filter(t => t.status === 'PENDING').length;
        
        // Recent agents table
        const recentAgentsTable = document.getElementById('recent-agents-table');
        recentAgentsTable.innerHTML = agents.slice(0, 5).map(agent => `
            <tr>
                <td>${agent.device_id}</td>
                <td>${renderStatus(agent.state)}</td>
                <td>${formatDate(agent.last_seen)}</td>
                <td>${agent.address || '-'}</td>
            </tr>
        `).join('');
        
        // Recent tasks table
        const recentTasksTable = document.getElementById('recent-tasks-table');
        recentTasksTable.innerHTML = tasks.slice(0, 5).map(task => `
            <tr>
                <td><code>${task.job_id.substring(0, 8)}...</code></td>
                <td><code>${task.agent_id.substring(0, 8)}...</code></td>
                <td>${renderStatus(task.status)}</td>
                <td>${formatDate(task.created_at)}</td>
            </tr>
        `).join('');
        
    } catch (error) {
        console.error('Failed to load dashboard:', error);
    }
}

// Agents
async function loadAgents() {
    try {
        const agents = await apiGet('/agents/list');
        const table = document.getElementById('agents-table');
        
        if (agents.length === 0) {
            table.innerHTML = `
                <tr>
                    <td colspan="6" class="empty-state">
                        <i class="fas fa-server"></i>
                        <p>No agents registered yet</p>
                    </td>
                </tr>
            `;
            return;
        }
        
        table.innerHTML = agents.map(agent => `
            <tr>
                <td><input type="checkbox" class="agent-checkbox" data-agent-id="${agent.id}"></td>
                <td>${agent.device_id}</td>
                <td>${renderStatus(agent.state)}</td>
                <td>${formatDate(agent.last_seen)}</td>
                <td>${agent.address || '-'}</td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="viewAgentDetails('${agent.id}')">
                        <i class="fas fa-eye"></i> View
                    </button>
                    ${agent.state === 'CONNECTED' ? `
                        <button class="btn btn-sm btn-success" onclick="showToolModal('${agent.id}')">
                            <i class="fas fa-play"></i> Tool
                        </button>
                    ` : ''}
                </td>
            </tr>
        `).join('');
        
        // Add checkbox listeners
        document.querySelectorAll('.agent-checkbox').forEach(cb => {
            cb.addEventListener('change', (e) => {
                const agentId = e.target.dataset.agentId;
                if (e.target.checked) {
                    selectedAgents.add(agentId);
                } else {
                    selectedAgents.delete(agentId);
                }
            });
        });
        
    } catch (error) {
        console.error('Failed to load agents:', error);
    }
}

async function viewAgentDetails(agentId) {
    try {
        console.log('Loading agent details for:', agentId);
        const details = await apiGet(`/agents/${agentId}`);
        console.log('Agent details response:', details);
        currentAgentId = agentId;
        
        // Validate response structure
        if (!details || !details.agent) {
            throw new Error('Invalid agent details response');
        }
        
        const content = document.getElementById('agent-details-content');
        content.innerHTML = `
            <div class="stats-grid" style="margin-bottom: 20px;">
                <div class="stat-card" style="padding: 16px;">
                    <div class="stat-info">
                        <h3 style="font-size: 16px;">${details.agent.device_id || 'Unknown'}</h3>
                        <p>Device ID</p>
                    </div>
                </div>
                <div class="stat-card" style="padding: 16px;">
                    <div class="stat-info">
                        <h3 style="font-size: 16px;">${renderStatus(details.agent.state)}</h3>
                        <p>Status</p>
                    </div>
                </div>
                <div class="stat-card" style="padding: 16px;">
                    <div class="stat-info">
                        <h3 style="font-size: 16px;">${formatDate(details.agent.last_seen)}</h3>
                        <p>Last Seen</p>
                    </div>
                </div>
            </div>
        `;
        
        // Load agent tasks
        try {
            const tasks = await apiGet(`/agents/${agentId}/tasks`);
            console.log('Agent tasks:', tasks);
            document.getElementById('agent-tasks-table').innerHTML = Array.isArray(tasks) && tasks.length > 0 
                ? tasks.map(task => `
                    <tr>
                        <td><code>${task.job_id ? task.job_id.substring(0, 8) : 'N/A'}...</code></td>
                        <td>${renderStatus(task.status)}</td>
                        <td>${task.result ? `<code>${task.result.substring(0, 50)}...</code>` : '-'}</td>
                        <td>${formatDate(task.created_at)}</td>
                    </tr>
                `).join('')
                : '<tr><td colspan="4" class="empty-state">No tasks assigned</td></tr>';
        } catch (taskErr) {
            console.error('Failed to load agent tasks:', taskErr);
            document.getElementById('agent-tasks-table').innerHTML = 
                '<tr><td colspan="4" class="empty-state">Failed to load tasks</td></tr>';
        }
        
        // Load agent services
        try {
            const services = details.services || [];
            console.log('Agent services:', services);
            document.getElementById('agent-services-table').innerHTML = services.length > 0
                ? services.map(service => `
                    <tr>
                        <td>${service.name || 'Unknown'}</td>
                        <td>${renderStatus(service.state)}</td>
                        <td>${formatDate(service.updated_at)}</td>
                    </tr>
                `).join('')
                : '<tr><td colspan="3" class="empty-state">No services registered</td></tr>';
        } catch (svcErr) {
            console.error('Failed to render services:', svcErr);
            document.getElementById('agent-services-table').innerHTML = 
                '<tr><td colspan="3" class="empty-state">Failed to load services</td></tr>';
        }
        
        // Load agent tools if connected
        if (details.agent.state === 'CONNECTED') {
            try {
                const tools = await apiGet(`/agents/${agentId}/tools`);
                console.log('Agent tools:', tools);
                currentAgentTools = Array.isArray(tools) ? tools : [];
                renderAgentTools(currentAgentTools);
            } catch (e) {
                console.error('Failed to load tools:', e);
                document.getElementById('agent-tools-list').innerHTML = 
                    '<div class="empty-state">Failed to load tools: ' + (e.message || 'Unknown error') + '</div>';
            }
        } else {
            document.getElementById('agent-tools-list').innerHTML = 
                '<div class="empty-state">Agent is not connected</div>';
        }
        
        showModal('agent-modal');
        // Activate first tab by default
        const firstTabBtn = document.querySelector('#agent-modal .tab-btn');
        if (firstTabBtn) {
            showAgentTab('tasks', firstTabBtn);
        }
        
    } catch (error) {
        console.error('Failed to load agent details:', error);
        showToast('Failed to load agent details: ' + (error.message || 'Unknown error'), 'error');
    }
}

function escapeHtml(str) {
    if (!str) return '';
    return str
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}

function renderAgentTools(tools) {
    const container = document.getElementById('agent-tools-list');
    if (!Array.isArray(tools) || tools.length === 0) {
        container.innerHTML = '<div class="empty-state">No tools available</div>';
        return;
    }
    
    container.innerHTML = tools.map(tool => {
        const toolName = escapeHtml(tool.name);
        const osTags = Array.isArray(tool.os) 
            ? tool.os.map(o => `<span class="os-tag">${escapeHtml(o)}</span>`).join('')
            : '';
        const actions = Array.isArray(tool.actions)
            ? tool.actions.map(action => {
                const actionName = escapeHtml(action.name);
                return `
                    <button class="tool-action-btn" onclick="quickExecuteTool('${toolName}', '${actionName}')">
                        ${actionName}
                    </button>
                `;
            }).join('')
            : '<span class="empty-state">No actions available</span>';
        
        return `
            <div class="tool-card">
                <h4>${toolName}</h4>
                <div class="os-tags">${osTags}</div>
                <div class="tool-actions">${actions}</div>
            </div>
        `;
    }).join('');
}

function showAgentTab(tab, clickedBtn) {
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
    
    if (clickedBtn) {
        clickedBtn.classList.add('active');
    }
    document.getElementById(`agent-${tab}-tab`).classList.add('active');
}

function showUnregisterModal() {
    if (selectedAgents.size === 0) {
        showToast('Please select agents to unregister', 'warning');
        return;
    }
    
    if (confirm(`Are you sure you want to unregister ${selectedAgents.size} agent(s)?`)) {
        // TODO: Implement batch unregister
        showToast('Unregister not implemented yet', 'info');
    }
}

// Jobs
async function loadJobs() {
    try {
        const jobs = await apiGet('/jobs/list');
        const table = document.getElementById('jobs-table');
        
        if (jobs.length === 0) {
            table.innerHTML = `
                <tr>
                    <td colspan="5" class="empty-state">
                        <i class="fas fa-tasks"></i>
                        <p>No jobs created yet</p>
                    </td>
                </tr>
            `;
            return;
        }
        
        table.innerHTML = jobs.map(job => `
            <tr>
                <td>${job.name}</td>
                <td><span class="badge badge-${job.type.toLowerCase()}">${job.type}</span></td>
                <td>${job.description || '-'}</td>
                <td><code>${job.target.substring(0, 30)}...</code></td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="viewJobDetails('${job.id}')">
                        <i class="fas fa-eye"></i>
                    </button>
                </td>
            </tr>
        `).join('');
        
    } catch (error) {
        console.error('Failed to load jobs:', error);
    }
}

function showCreateJobModal() {
    showModal('create-job-modal');
}

async function handleCreateJob(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const data = {
        name: formData.get('name'),
        type: formData.get('type'),
        description: formData.get('description'),
        target: formData.get('target')
    };
    
    try {
        await apiPost('/jobs/create', data);
        closeModal('create-job-modal');
        e.target.reset();
        loadJobs();
        showToast('Job created successfully', 'success');
    } catch (error) {
        // Error already shown by apiPost
    }
}

// Tasks
async function loadTasks() {
    try {
        const tasks = await apiGet('/tasks/list');
        const table = document.getElementById('tasks-table');
        
        if (tasks.length === 0) {
            table.innerHTML = `
                <tr>
                    <td colspan="6" class="empty-state">
                        <i class="fas fa-clipboard-list"></i>
                        <p>No tasks assigned yet</p>
                    </td>
                </tr>
            `;
            return;
        }
        
        table.innerHTML = tasks.map(task => `
            <tr>
                <td><code>${task.id.substring(0, 8)}...</code></td>
                <td><code>${task.job_id.substring(0, 8)}...</code></td>
                <td><code>${task.agent_id.substring(0, 8)}...</code></td>
                <td>${renderStatus(task.status)}</td>
                <td>${task.result ? `<code>${task.result.substring(0, 30)}...</code>` : '-'}</td>
                <td>${formatDate(task.created_at)}</td>
            </tr>
        `).join('');
        
    } catch (error) {
        console.error('Failed to load tasks:', error);
    }
}

async function showAssignTaskModal() {
    try {
        const [agents, jobs] = await Promise.all([
            apiGet('/agents/list'),
            apiGet('/jobs/list')
        ]);
        
        const agentSelect = document.getElementById('task-agent-select');
        const jobSelect = document.getElementById('task-job-select');
        
        agentSelect.innerHTML = agents.map(a => 
            `<option value="${a.id}">${a.device_id} (${a.state})</option>`
        ).join('');
        
        jobSelect.innerHTML = jobs.map(j => 
            `<option value="${j.id}">${j.name} (${j.type})</option>`
        ).join('');
        
        showModal('assign-task-modal');
    } catch (error) {
        showToast('Failed to load data', 'error');
    }
}

async function handleAssignTask(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const data = {
        agent_id: formData.get('agent_id'),
        job_id: formData.get('job_id')
    };
    
    try {
        await apiPost('/tasks/assign', data);
        closeModal('assign-task-modal');
        e.target.reset();
        loadTasks();
        showToast('Task assigned successfully', 'success');
    } catch (error) {
        // Error already shown
    }
}

// Queries
async function loadQueries() {
    try {
        const queries = await apiGet('/queries/list');
        const table = document.getElementById('queries-table');
        
        if (queries.length === 0) {
            table.innerHTML = `
                <tr>
                    <td colspan="5" class="empty-state">
                        <i class="fas fa-database"></i>
                        <p>No queries created yet</p>
                    </td>
                </tr>
            `;
            return;
        }
        
        table.innerHTML = queries.map(query => `
            <tr>
                <td>${query.name}</td>
                <td>${query.description || '-'}</td>
                <td>${query.platform || 'All'}</td>
                <td><code>${query.sql.substring(0, 40)}...</code></td>
                <td>
                    <button class="btn btn-sm btn-success" onclick="showRunQueryModal('${query.id}')">
                        <i class="fas fa-play"></i> Run
                    </button>
                    <button class="btn btn-sm btn-danger" onclick="deleteQuery('${query.id}')">
                        <i class="fas fa-trash"></i>
                    </button>
                </td>
            </tr>
        `).join('');
        
    } catch (error) {
        console.error('Failed to load queries:', error);
    }
}

function showCreateQueryModal() {
    showModal('create-query-modal');
}

async function handleCreateQuery(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const data = {
        name: formData.get('name'),
        description: formData.get('description'),
        sql: formData.get('sql'),
        platform: formData.get('platform')
    };
    
    try {
        await apiPost('/queries/create', data);
        closeModal('create-query-modal');
        e.target.reset();
        loadQueries();
        showToast('Query created successfully', 'success');
    } catch (error) {
        // Error already shown
    }
}

async function showRunQueryModal(queryId) {
    try {
        const agents = await apiGet('/agents/list');
        const container = document.getElementById('run-query-agents');
        
        document.getElementById('run-query-id').value = queryId;
        
        container.innerHTML = agents.map(agent => `
            <label>
                <input type="checkbox" name="agent_ids" value="${agent.id}">
                ${agent.device_id} (${agent.state})
            </label>
        `).join('');
        
        showModal('run-query-modal');
    } catch (error) {
        showToast('Failed to load agents', 'error');
    }
}

async function handleRunQuery(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const agentIds = formData.getAll('agent_ids');
    
    const data = {
        query_id: formData.get('query_id'),
        agent_ids: agentIds
    };
    
    try {
        await apiPost('/queries/run', data);
        closeModal('run-query-modal');
        e.target.reset();
        showToast('Query execution started', 'success');
        
        // Switch to executions page
        document.querySelector('[data-page="executions"]').click();
    } catch (error) {
        // Error already shown
    }
}

async function deleteQuery(queryId) {
    if (!confirm('Are you sure you want to delete this query?')) return;
    
    try {
        await apiDelete(`/queries/${queryId}`);
        loadQueries();
        showToast('Query deleted', 'success');
    } catch (error) {
        // Error already shown
    }
}

// Executions
async function loadExecutions() {
    try {
        const executions = await apiGet('/query-executions/list');
        const table = document.getElementById('executions-table');
        
        if (!Array.isArray(executions) || executions.length === 0) {
            table.innerHTML = `
                <tr>
                    <td colspan="5" class="empty-state">
                        <i class="fas fa-history"></i>
                        <p>No query executions yet</p>
                    </td>
                </tr>
            `;
            return;
        }
        
        table.innerHTML = executions.map(exec => `
            <tr>
                <td>${exec.query?.name || 'Unknown'}</td>
                <td>${renderStatus(exec.status)}</td>
                <td>${formatDate(exec.created_at)}</td>
                <td>${exec.completed_at ? formatDate(exec.completed_at) : '-'}</td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="viewExecutionDetails('${exec.id}')">
                        <i class="fas fa-eye"></i> View
                    </button>
                </td>
            </tr>
        `).join('');
        
    } catch (error) {
        console.error('Failed to load executions:', error);
        const table = document.getElementById('executions-table');
        table.innerHTML = `
            <tr>
                <td colspan="5" class="empty-state">
                    <i class="fas fa-exclamation-circle"></i>
                    <p>Failed to load executions</p>
                </td>
            </tr>
        `;
    }
}

async function viewExecutionDetails(executionId) {
    try {
        console.log('Loading execution details for:', executionId);
        const response = await apiGet(`/query-executions/${executionId}`);
        console.log('Execution details response:', response);
        
        // Handle both {execution, results} wrapper and direct response
        const execution = response.execution || response;
        const results = response.results || [];
        
        if (!execution) {
            throw new Error('Invalid execution response');
        }
        
        document.getElementById('execution-details-content').innerHTML = `
            <div class="stats-grid" style="margin-bottom: 20px;">
                <div class="stat-card" style="padding: 16px;">
                    <div class="stat-info">
                        <h3 style="font-size: 16px;">${execution.query?.name || 'Unknown'}</h3>
                        <p>Query</p>
                    </div>
                </div>
                <div class="stat-card" style="padding: 16px;">
                    <div class="stat-info">
                        <h3 style="font-size: 16px;">${renderStatus(execution.status)}</h3>
                        <p>Status</p>
                    </div>
                </div>
            </div>
            <pre style="margin-bottom: 20px;"><code>${execution.query?.sql || 'N/A'}</code></pre>
        `;
        
        // Render execution results
        const resultsTable = document.getElementById('execution-results-table');
        if (Array.isArray(results) && results.length > 0) {
            resultsTable.innerHTML = results.map(result => `
                <tr>
                    <td>${result.agent?.device_id || result.agent_id?.substring(0, 8) || 'Unknown'}</td>
                    <td>${renderStatus(result.status)}</td>
                    <td>${result.result_json ? `<pre><code>${result.result_json.substring(0, 200)}...</code></pre>` : '-'}</td>
                    <td>${result.error || '-'}</td>
                </tr>
            `).join('');
        } else {
            resultsTable.innerHTML = `
                <tr>
                    <td colspan="4" class="empty-state">No results available yet</td>
                </tr>
            `;
        }
        
        showModal('execution-modal');
    } catch (error) {
        console.error('Failed to load execution details:', error);
        showToast('Failed to load execution details: ' + (error.message || 'Unknown error'), 'error');
    }
}

// Tools
async function showToolModal(agentId) {
    currentAgentId = agentId;
    
    try {
        const tools = await apiGet(`/agents/${agentId}/tools`);
        currentAgentTools = tools;
        
        const toolSelect = document.getElementById('tool-name-select');
        toolSelect.innerHTML = tools.map(t => 
            `<option value="${t.name}">${t.name}</option>`
        ).join('');
        
        document.getElementById('tool-agent-id').value = agentId;
        
        updateToolActions();
        showModal('tool-modal');
    } catch (error) {
        showToast('Failed to load tools', 'error');
    }
}

function updateToolActions() {
    const toolName = document.getElementById('tool-name-select').value;
    const tool = currentAgentTools.find(t => t.name === toolName);
    
    const actionSelect = document.getElementById('tool-action-select');
    if (tool) {
        actionSelect.innerHTML = tool.actions.map(a => 
            `<option value="${a.name}">${a.name}</option>`
        ).join('');
    }
}

function quickExecuteTool(toolName, actionName) {
    // Ensure agent ID is set
    if (!currentAgentId) {
        showToast('No agent selected', 'error');
        return;
    }
    document.getElementById('tool-agent-id').value = currentAgentId;
    document.getElementById('tool-name-select').value = toolName;
    updateToolActions();
    document.getElementById('tool-action-select').value = actionName;
    closeModal('agent-modal');
    showModal('tool-modal');
}

async function handleExecuteTool(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const optionsStr = formData.get('tool_options');
    const options = optionsStr ? optionsStr.split(',').map(s => s.trim()) : [];
    
    const data = {
        agent_id: formData.get('agent_id'),
        tool_name: formData.get('tool_name'),
        tool_action: formData.get('tool_action'),
        tool_options: options
    };
    
    try {
        await apiPost('/tools/execute', data);
        closeModal('tool-modal');
        e.target.reset();
        showToast('Tool execution started', 'success');
    } catch (error) {
        // Error already shown
    }
}

// Utility Functions
function renderStatus(status) {
    const statusMap = {
        'CONNECTED': 'badge-connected',
        'DISCONNECTED': 'badge-disconnected',
        'PENDING': 'badge-pending',
        'RUNNING': 'badge-running',
        'COMPLETED': 'badge-completed',
        'FAILED': 'badge-failed',
        'COMMAND': 'badge-info',
        'SCRIPT': 'badge-warning'
    };
    
    const badgeClass = statusMap[status] || 'badge-pending';
    return `<span class="badge ${badgeClass}">${status}</span>`;
}

function formatDate(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleString();
}

function showModal(modalId) {
    document.getElementById(modalId).classList.add('active');
}

function closeModal(modalId) {
    document.getElementById(modalId).classList.remove('active');
}

function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    
    const icons = {
        success: 'check-circle',
        error: 'exclamation-circle',
        warning: 'exclamation-triangle',
        info: 'info-circle'
    };
    
    toast.innerHTML = `
        <i class="fas fa-${icons[type]}"></i>
        <span>${message}</span>
    `;
    
    container.appendChild(toast);
    
    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateX(100%)';
        setTimeout(() => toast.remove(), 300);
    }, 4000);
}

// Close modals on outside click
window.onclick = function(event) {
    if (event.target.classList.contains('modal')) {
        event.target.classList.remove('active');
    }
}
