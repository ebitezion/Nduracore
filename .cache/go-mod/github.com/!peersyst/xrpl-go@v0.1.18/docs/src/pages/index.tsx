import type {ReactNode} from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import {LinkButton} from '../components/LinkButton';
import config from '@site/docusaurus.config';

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`Hello from ${siteConfig.title}`}
      description="Description will go into a meta tag in <head />">
      <main>
        <section className="hero-section">
          <img 
            src={`${config.baseUrl}/img/xrpl-go-logo.png`}
            alt="XRPL GO Logo"
            className="hero-logo"
          />
          <h1 className="hero-title">
            XRPL GO
          </h1>
          <p className="hero-description">
            A comprehensive Go library for interacting with the XRP Ledger. 
            Built with performance and developer experience in mind, XRPL Go provides 
            all the tools needed to build robust applications on the XRPL ecosystem.
          </p>
          <LinkButton href={`${config.baseUrl}/docs/intro`}>
            Getting Started
          </LinkButton>
        </section>
      </main>
    </Layout>
  );
}
