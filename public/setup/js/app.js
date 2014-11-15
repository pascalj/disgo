(function() {
  var app = angular.module('setupApp', ['ngResource'])


  app.controller('DatabaseController', function($scope, $resource, $http) {
    var Database = $resource('/setup/database', {}, {test: {method:'GET'}});

    $scope.database = {
      driver: 'sqlite3'
    }

    $scope.changeDriver = function(driver) {
      $scope.database.driver = driver;
      $http.get('/setup/database', {params: $scope.database}).success(function(result) {
        console.log(result)
        $scope.valid = result;
      });
    }

    $scope.submit = function() {
      console.log(setup)
    }

    $scope.isActiveTab = function(driver) {
      return $scope.database.driver == driver;
    }
  })
})()
