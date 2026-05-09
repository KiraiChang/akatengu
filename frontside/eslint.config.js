import eslint from '@eslint/js';
import tseslint from '@typescript-eslint/eslint-plugin';
import tsparser from '@typescript-eslint/parser';
import sveltePlugin from 'eslint-plugin-svelte';
import svelteParser from 'svelte-eslint-parser';
import globals from 'globals';

export default [
    // 忽略的目錄
    {
        ignores: ['node_modules/**', 'dist/**', '.svelte-kit/**', 'build/**'],
    },

    // TypeScript 檔案（含 .svelte.ts rune 模組）
    {
        files: ['**/*.ts'],
        languageOptions: {
            parser: tsparser,
            parserOptions: {
                project: './tsconfig.app.json',
            },
            globals: {
                ...globals.browser,
                ...globals.node,
                // Svelte 5 runes：在 .svelte.ts 模組中為編譯器巨集，非真實全域
                $state: 'readonly',
                $derived: 'readonly',
                $effect: 'readonly',
                $props: 'readonly',
                $bindable: 'readonly',
                $inspect: 'readonly',
                $host: 'readonly',
            },
        },
        plugins: {
            '@typescript-eslint': tseslint,
        },
        rules: {
            ...eslint.configs.recommended.rules,
            ...tseslint.configs.recommended.rules,
            '@typescript-eslint/no-unused-vars': ['warn', { varsIgnorePattern: '^_', argsIgnorePattern: '^_' }],
            '@typescript-eslint/no-explicit-any': 'warn',
        },
    },

    // Svelte 檔案
    {
        files: ['**/*.svelte'],
        languageOptions: {
            parser: svelteParser,
            parserOptions: {
                parser: tsparser, // svelte 內的 <script lang="ts"> 用 ts parser
            },
            globals: {
                ...globals.browser,
            },
        },
        plugins: {
            svelte: sveltePlugin,
            '@typescript-eslint': tseslint,
        },
        rules: {
            ...sveltePlugin.configs.recommended.rules,
            'no-unused-vars': 'off', // 用 @typescript-eslint 版本取代
            '@typescript-eslint/no-unused-vars': ['warn', { varsIgnorePattern: '^_', argsIgnorePattern: '^_' }],
        },
    },
];