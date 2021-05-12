module.exports = {
	root: true,
	env: {
		node: true,
	},
	extends: [
		'plugin:vue/essential',
		'plugin:vue/recommended',
		'eslint:recommended',
		'@vue/typescript/recommended'
	],
	parserOptions: {
		ecmaVersion: 2020,
	},
	ignorePatterns: ['src/rpc/protocol/*.js', 'src/rpc/protocol/*.d.ts'],
	rules: {
		'no-console': 'off',
		'no-debugger': 'off',
		'no-tabs': 0,
		semi: 0,
		indent: [2, 'tab', {
			SwitchCase: 1
		}],
		'no-unused-vars': 1,
		'vue/no-unused-components': 1,
		'no-useless-escape': 0,
		'no-case-declarations': 1,
		quotes: [2, 'single', {
			avoidEscape: true,
			allowTemplateLiterals: true
		}],
		'keyword-spacing': [2, {
			before: true,
			after: true
		}],
		'comma-spacing': [2, {
			before: false,
			after: true
		}],
		'object-curly-spacing': [2, 'always'],
		'array-bracket-spacing': [1, 'never'],
		'no-trailing-spaces': 2,
		'space-before-function-paren': 0,
		'block-spacing': [2, 'always'],
		'key-spacing': [2, {
			afterColon: true
		}],
		'vue/html-indent': [1, 'tab'],
		'vue/max-attributes-per-line': [1, {
			singleline: 2,
			multiline: {
				allowFirstLine: true
			}
		}],
		'vue/singleline-html-element-content-newline': 0,
		'vue/html-closing-bracket-newline': 0,
		'vue/valid-v-slot': 0,
		'consistent-return': 2,
		'no-else-return': 1,
		'max-len': [1, {
			code: 120,
			ignoreComments: true,
			ignoreTrailingComments: true,
			ignoreUrls: true,
			ignoreStrings: true,
			ignoreTemplateLiterals: true,
			ignoreRegExpLiterals: true
		}],
		'func-names': 0,
		curly: 0,
		'new-cap': 0,
		'prefer-destructuring': ['error', {
			array: false,
			object: true
		}],
		'quote-props': ['error', 'as-needed'],
		'operator-linebreak': ['error', 'before']
	}
};
