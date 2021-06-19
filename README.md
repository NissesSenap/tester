# Tester

A test controller for flagger starting test deployments.
Flagger will use a rest API to create requests to start tests.

## Commands

### kubebuilder

We won't use this in the end.

```shell
kubebuilder init --domain flagger.app --repo flagger.app/tester
kubebuilder create api --group tester --version v1alpha1 --kind Tester
```

## Questions

- Do we see any issues with using fluxcd/pkg in a flagger related project?
