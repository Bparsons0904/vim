import { Component } from "solid-js";
import { Card } from "@components/common/ui/Card/Card";
import styles from "./ProcessInfo.module.scss";

export const ProcessInfo: Component = () => {
  return (
    <Card class={styles.processInfoCard}>
      <h2>How the Test Works</h2>
      <div class={styles.processSteps}>
        <div class={styles.processStep}>
          <div class={styles.stepNumber}>1</div>
          <div class={styles.stepContent}>
            <h3>CSV Generation</h3>
            <p>
              Generate test data with 150 columns including various date
              fields that can be one of 10 different data types
            </p>
          </div>
        </div>
        <div class={styles.processStep}>
          <div class={styles.stepNumber}>2</div>
          <div class={styles.stepContent}>
            <h3>Data Parsing & Validation</h3>
            <p>
              Parse the CSV data and identify 25 columns of interest.
              Validate and normalize date fields to standard timestamp
              format
            </p>
          </div>
        </div>
        <div class={styles.processStep}>
          <div class={styles.stepNumber}>3</div>
          <div class={styles.stepContent}>
            <h3>Database Insertion</h3>
            <p>
              Insert the processed columns as strings into the database
              using the selected insertion method
            </p>
          </div>
        </div>
      </div>
    </Card>
  );
};