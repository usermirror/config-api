import Ember from 'ember';
import ControlGroupError from 'vault/lib/control-group-error';

const { inject } = Ember;
export default Ember.Route.extend({
  controlGroup: inject.service(),
  routing: inject.service('router'),
  namespaceService: inject.service('namespace'),

  actions: {
    willTransition() {
      window.scrollTo(0, 0);
    },
    error(error, transition) {
      let controlGroup = this.get('controlGroup');
      if (error instanceof ControlGroupError) {
        return controlGroup.handleError(error, transition);
      }
      if (error.path === '/v1/sys/wrapping/unwrap') {
        controlGroup.unmarkTokenForUnwrap();
      }

      let router = this.get('routing');
      let errorURL = transition.intent.url;
      let { name, contexts, queryParams } = transition.intent;

      // If the transition is internal to Ember, we need to generate the URL
      // from the route parameters ourselves
      if (!errorURL) {
        try {
          errorURL = router.urlFor(name, ...(contexts || []), { queryParams });
        } catch (e) {
          // If this fails, something weird is happening with URL transitions
          errorURL = null;
        }
      }
      // because we're using rootURL, we need to trim this from the front to get
      // the ember-routeable url
      if (errorURL) {
        errorURL = errorURL.replace('/ui', '');
      }

      error.errorURL = errorURL;

      // if we have queryParams, update the namespace so that the observer can fire on the controller
      if (queryParams) {
        this.controllerFor('vault.cluster').set('namespaceQueryParam', queryParams.namespace || '');
      }

      // Assuming we have a URL, push it into browser history and update the
      // location bar for the user
      if (errorURL) {
        router.get('location').setURL(errorURL);
      }

      return true;
    },
  },
});
