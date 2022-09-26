import { IngressOptions } from "ark/plugins/@mantl/sre/shared";


/**
 * Helper function that generate a default k8s ingress object for NGINX controller for
   the '/' path given a service name and a set of urls
 * @param {string} serviceName - The k8s service name used as backend in the ingress path
 * @param {string[]} urls - The list of urls to be used as hosts for the nginx ingress
 * @param {number} [servicePort=80] - The port nubmer of the service used as backend in the ingress path
 * @returns {IngressOptions} - An ingress option full compatible with k8s ingress 
 * @example
 *  const i = generateNginxDefaultIngress('console-web', ['pr-3323-console.prod.mantl.com']) 
 *  // returns 
 *  {
 *    name: 'console-web',
 *    rules: [{
 *      host: 'pr-3323-console.eph.mantl.dev',
 *      http: {
 *        paths: [{
 *          path: '/',
 *          backend: {
 *            service: {
 *              name: 'console-web',
 *              port: {
 *                number: 80, 
 *              }
 *            }
 *          }
 *        }]
 *      }
 *    }],
 *    annotations: {
 *      'kubernetes.io/ingress.class': 'nginx',
 *    },
 *  }
**/
export const generateNginxDefaultIngress = (serviceName: string, urls: string[], servicePort: number = 80):
  IngressOptions => ({
    name: serviceName,
    rules: urls?.map((url: string) => ({
      host: url,
      http: {
        paths: [{
          path: '/',
          pathType: 'ImplementationSpecific',
          backend: {
            service: {
              name: serviceName,
              port: {
                number: servicePort,
              }
            }
          }
        }]
      }
    })),
    annotations: {
      'kubernetes.io/ingress.class': 'nginx',
    },
  })

/**
 * helper function that add annotations to a given ingress object
 * @param {IngressOptions} ingress - The ingress object
 * @param {{[key:string]: string}} - An key/value string dictionary with the annotations
 * @param {IngressOptions} A copy of the passed ingress object extended with the passed annotations
 * @example
 *  const ingress = generateNginxDefaultIngress('console-web', ['pr-3323-console.prod.mantl.com']) 
 *  const i = addAnnotationsToIngress(ingress, { 'nginx.org/client-max-body-size': '10M' })
 *  // returns 
 *  {
 *    name: 'console-web',
 *    rules: [{
 *      host: 'pr-3323-console.eph.mantl.dev',
 *      http: {
 *        paths: [{
 *          path: '/',
 *          backend: {
 *            service: {
 *              name: 'console-web',
 *              port: {
 *                number: 80, 
 *              }
 *            }
 *          }
 *        }]
 *      }
 *    }],
 *    annotations: {
 *      'kubernetes.io/ingress.class': 'nginx',
        'nginx.org/client-max-body-size': '10M',
 *    },
 *  }
*/
export const addAnnotationsToIngress = (ingress: IngressOptions, annotations: { [key: string]: string }):
  IngressOptions => {
  const r = { ...ingress }
  r.annotations = {
    ...r.annotations,
    ...annotations,
  }
  return r
}
