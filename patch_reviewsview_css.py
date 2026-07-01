with open('frontend/src/views/reviews/ReviewsView.vue', 'r') as f:
    content = f.read()

css_addition = '''
.modern-pills {
  display: inline-flex;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 9999px;
  padding: 4px;
}
.pill-btn {
  padding: 6px 14px;
  border-radius: 9999px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted);
  transition: all 0.2s ease;
  background: transparent;
  border: none;
  cursor: pointer;
}
.pill-btn:hover {
  color: var(--text-main);
}
.pill-btn.active {
  background: var(--primary-color, #6366f1);
  color: #fff;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.4);
}
'''

if '.modern-pills' not in content:
    content = content.replace('</style>', css_addition + '\n</style>')
    with open('frontend/src/views/reviews/ReviewsView.vue', 'w') as f:
        f.write(content)

