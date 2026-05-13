import { Link } from '@umijs/max';
import styles from './index.less';

interface PageLink {
  title: string;
  path: string;
  description: string;
}

interface LearningItem {
  title: string;
  content: string;
}

const pageLinks: PageLink[] = [
  {
    title: 'API 文档',
    path: '/apidocs',
    description: '查看当前前端已经配置好的 API 文档页面。',
  },
  {
    title: '登录页',
    path: '/login',
    description: '后续学习表单、输入框、按钮和提交事件时会用到。',
  },
];

const learningItems: LearningItem[] = [
  {
    title: '1. Component 组件',
    content: 'React 页面由组件组成。这个 HomePage 本身就是一个组件。',
  },
  {
    title: '2. JSX',
    content:
      'JSX 让我们在 TypeScript 中写类似 HTML 的结构，最后由 React 渲染成页面。',
  },
  {
    title: '3. Props',
    content:
      'Props 是组件的输入。后面我们会把重复的卡片抽成组件，再用 props 传入标题和内容。',
  },
  {
    title: '4. List Rendering',
    content:
      '这里的学习内容来自数组，并用 map 渲染出来，这是 React 中很常见的写法。',
  },
];

const HomePage = () => {
  return (
    <main className={styles.page}>
      <section className={styles.hero}>
        <p className={styles.eyebrow}>LinkFlow Web</p>
        <h1>React 学习首页</h1>
        <p className={styles.summary}>
          这个页面先作为学习入口：上面放项目页面链接，下面放 React
          基础概念。每一步都保持小改动，方便你对照代码理解。
        </p>
      </section>

      <section className={styles.section} aria-labelledby="page-links-title">
        <h2 id="page-links-title">项目页面链接</h2>
        <div className={styles.linkGrid}>
          {pageLinks.map((item) => (
            <Link key={item.path} to={item.path} className={styles.linkCard}>
              <span>{item.title}</span>
              <p>{item.description}</p>
            </Link>
          ))}
        </div>
      </section>

      <section className={styles.section} aria-labelledby="learning-title">
        <h2 id="learning-title">React 基础学习内容</h2>
        <div className={styles.learningGrid}>
          {learningItems.map((item) => (
            <article key={item.title} className={styles.learningCard}>
              <h3>{item.title}</h3>
              <p>{item.content}</p>
            </article>
          ))}
        </div>
      </section>

      <section className={styles.codeSection} aria-labelledby="code-title">
        <h2 id="code-title">这一步重点看什么</h2>
        <pre>
          <code>{`const items = ['Component', 'JSX', 'Props'];

items.map((item) => (
  <div key={item}>{item}</div>
));`}</code>
        </pre>
      </section>
    </main>
  );
};

export default HomePage;
