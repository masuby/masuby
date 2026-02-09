// Transparency Page JavaScript

// Tab switching
document.querySelectorAll('.tab-btn').forEach(function(btn) {
    btn.addEventListener('click', function() {
        document.querySelectorAll('.tab-btn').forEach(function(b) { b.classList.remove('active'); });
        document.querySelectorAll('.tab-content').forEach(function(c) { c.classList.remove('active'); });

        btn.classList.add('active');
        document.getElementById(btn.dataset.tab).classList.add('active');
    });
});

// Load data on page load
document.addEventListener('DOMContentLoaded', function() {
    loadFormulas();
    loadDataFlow();
    loadLinkages();
    loadAPIDoc();
});

function loadFormulas() {
    fetch('/api/v1/transparency/formulas')
        .then(function(response) { return response.json(); })
        .then(function(data) {
            var container = document.getElementById('formulas-container');
            if (data.success && data.data) {
                var html = '';
                data.data.forEach(function(formula) {
                    html += '<div class="formula-card">';
                    html += '<h3>' + formula.name + '</h3>';
                    html += '<div class="formula-display">' + formula.formula + '</div>';
                    html += '<p>' + formula.description + '</p>';
                    html += '<div class="formula-meta">';
                    html += '<div class="meta-item"><strong>Inputs</strong>' + formula.inputs.join(', ') + '</div>';
                    html += '<div class="meta-item"><strong>Output</strong>' + formula.output + '</div>';
                    html += '<div class="meta-item"><strong>Example</strong><code>' + formula.example + '</code></div>';
                    html += '</div></div>';
                });
                container.innerHTML = html;
            }
        })
        .catch(function(err) {
            console.error('Failed to load formulas:', err);
        });
}

function loadDataFlow() {
    fetch('/api/v1/transparency/dataflow')
        .then(function(response) { return response.json(); })
        .then(function(data) {
            var container = document.getElementById('dataflow-container');
            if (data.success && data.data) {
                var flowHtml = '<div class="flow-diagram">';
                data.data.forEach(function(step, index) {
                    if (index > 0) {
                        flowHtml += '<span class="flow-arrow">&#8594;</span>';
                    }
                    flowHtml += '<div class="flow-step">';
                    flowHtml += '<div class="flow-step-number">' + step.step + '</div>';
                    flowHtml += '<div><strong>' + step.name + '</strong></div>';
                    flowHtml += '</div>';
                });
                flowHtml += '</div>';

                flowHtml += '<div style="margin-top: 2rem;">';
                data.data.forEach(function(step) {
                    flowHtml += '<div class="formula-card">';
                    flowHtml += '<h3>Step ' + step.step + ': ' + step.name + '</h3>';
                    flowHtml += '<p>' + step.description + '</p>';
                    flowHtml += '<div class="formula-meta">';
                    flowHtml += '<div class="meta-item"><strong>Input</strong>' + step.input + '</div>';
                    flowHtml += '<div class="meta-item"><strong>Output</strong>' + step.output + '</div>';
                    flowHtml += '<div class="meta-item"><strong>Related Sheets</strong>' + step.sheets.join(', ') + '</div>';
                    flowHtml += '</div></div>';
                });
                flowHtml += '</div>';

                container.innerHTML = flowHtml;
            }
        })
        .catch(function(err) {
            console.error('Failed to load data flow:', err);
        });
}

function loadLinkages() {
    fetch('/api/v1/transparency/linkages')
        .then(function(response) { return response.json(); })
        .then(function(data) {
            var container = document.getElementById('linkages-container');
            if (data.success && data.data) {
                var html = '<table class="linkage-table">';
                html += '<thead><tr><th>Source</th><th>Target</th><th>Link Type</th><th>Description</th></tr></thead>';
                html += '<tbody>';
                data.data.forEach(function(link) {
                    html += '<tr>';
                    html += '<td>' + link.source + '</td>';
                    html += '<td>' + link.target + '</td>';
                    html += '<td><span class="link-type ' + link.link_type + '">' + link.link_type.replace('_', ' ') + '</span></td>';
                    html += '<td>' + link.description + '</td>';
                    html += '</tr>';
                });
                html += '</tbody></table>';

                html += '<div class="formula-card" style="margin-top: 2rem;">';
                html += '<h3>Linkage Types Explained</h3>';
                html += '<ul style="line-height: 2;">';
                html += '<li><span class="link-type data_input">data input</span> - Raw data entered by committees</li>';
                html += '<li><span class="link-type data_flow">data flow</span> - Data transformation between processing stages</li>';
                html += '<li><span class="link-type reference">reference</span> - Lookup values and validation rules</li>';
                html += '<li><span class="link-type aggregation">aggregation</span> - Combining values using formulas</li>';
                html += '<li><span class="link-type visualization">visualization</span> - Data displayed on dashboards and maps</li>';
                html += '</ul></div>';

                container.innerHTML = html;
            }
        })
        .catch(function(err) {
            console.error('Failed to load linkages:', err);
        });
}

function loadAPIDoc() {
    fetch('/api/v1/transparency/api')
        .then(function(response) { return response.json(); })
        .then(function(data) {
            var container = document.getElementById('api-container');
            if (data.success && data.data) {
                var doc = data.data;
                var html = '<div class="formula-card">';
                html += '<h3>API Information</h3>';
                html += '<p><strong>Version:</strong> ' + doc.version + '</p>';
                html += '<p><strong>Base URL:</strong> <code>' + doc.base_url + '</code></p>';
                html += '</div>';

                html += '<h3 style="margin: 2rem 0 1rem;">Endpoints</h3>';
                doc.endpoints.forEach(function(ep) {
                    html += '<div class="api-endpoint">';
                    html += '<span class="api-method ' + ep.method.toLowerCase() + '">' + ep.method + '</span>';
                    html += '<span class="api-path">' + doc.base_url + ep.path + '</span>';
                    html += '<p style="margin-top: 0.5rem; color: #6b7280;">' + ep.description + '</p>';
                    if (ep.auth) {
                        html += '<span style="font-size: 0.875rem; color: #b45309;">Requires Authentication</span>';
                    } else {
                        html += '<span style="font-size: 0.875rem; color: #15803d;">Public</span>';
                    }
                    if (ep.roles) {
                        html += '<span style="font-size: 0.875rem; color: #7c3aed; margin-left: 1rem;">Roles: ' + ep.roles.join(', ') + '</span>';
                    }
                    html += '</div>';
                });

                html += '<div class="formula-card" style="margin-top: 2rem;">';
                html += '<h3>Authentication</h3>';
                html += '<p>The API uses JWT (JSON Web Tokens) for authentication.</p>';
                html += '<ol style="line-height: 2;">';
                html += '<li>Obtain a token by posting credentials to <code>/api/v1/auth/login</code></li>';
                html += '<li>Include the token in subsequent requests:<br><code>Authorization: Bearer &lt;your-token&gt;</code></li>';
                html += '<li>Tokens expire after 24 hours</li>';
                html += '</ol></div>';

                html += '<div class="formula-card">';
                html += '<h3>Response Format</h3>';
                html += '<div class="formula-display">';
                html += '{\n    "success": true,\n    "message": "Optional message",\n    "data": { ... },\n    "error": "Error message if success is false"\n}';
                html += '</div></div>';

                container.innerHTML = html;
            }
        })
        .catch(function(err) {
            console.error('Failed to load API doc:', err);
        });
}
